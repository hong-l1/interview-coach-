import { INTERVIEW_API } from '@/config/api';
import { readUserIDFromToken } from '@/services/api/client';

export type InterviewMode = 'campus' | 'social' | 'special';

export interface ConversationItem {
  type: 'question' | 'answer';
  content: string;
  index?: number;
  timestamp: number;
}

export interface ActiveInterviewSession {
  mode: InterviewMode;
  sessionId: string;
  recordId?: number;
  startedAt: number;
  params: Record<string, any>;
  questionIndex: number;
  questionText: string;
  answeredCount: number;
  elapsed: number;
  phase?: string;
  conversationHistory: ConversationItem[];
}

export interface SessionInfoResponse {
  session: {
    session_id: string;
    record_id: number;
    status: string;
    start_time: number;
  };
  current_question_index?: number;
  current_question_text?: string;
  answered_count?: number;
  total_count?: number;
  elapsed_time?: number;
  metadata?: Record<string, string>;
}

const STORAGE_PREFIX = 'activeInterviewSession:';

export function getInterviewStorageKey(mode: InterviewMode) {
  return `${STORAGE_PREFIX}${mode}`;
}

export function loadActiveInterviewSession(mode: InterviewMode): ActiveInterviewSession | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(getInterviewStorageKey(mode));
    if (!raw) return null;
    return JSON.parse(raw) as ActiveInterviewSession;
  } catch {
    return null;
  }
}

export function saveActiveInterviewSession(mode: InterviewMode, session: ActiveInterviewSession) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(getInterviewStorageKey(mode), JSON.stringify(session));
}

export function clearActiveInterviewSession(mode: InterviewMode) {
  if (typeof window === 'undefined') return;
  window.localStorage.removeItem(getInterviewStorageKey(mode));
}

export function getAuthToken() {
  if (typeof window === 'undefined') return '';
  return window.localStorage.getItem('token') || '';
}

export async function fetchSessionInfo(sessionId: string, token: string): Promise<SessionInfoResponse> {
  const response = await fetch(`${INTERVIEW_API.SESSION_INFO}?session_id=${encodeURIComponent(sessionId)}`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    mode: 'cors',
  });

  if (!response.ok) {
    throw new Error(`session_info_${response.status}`);
  }

  return response.json();
}

export function buildSseRequest(token: string, body: Record<string, any>, resume = false, signal?: AbortSignal) {
  const url = resume ? INTERVIEW_API.RESUME_STREAM : INTERVIEW_API.START_STREAM;
  let userID = '';
  if (typeof window !== 'undefined') {
    try {
      const user = JSON.parse(window.localStorage.getItem('user') || 'null');
      userID = String(user?.id || user?.user_id || user?.ID || readUserIDFromToken(token) || '');
    } catch {
      userID = '';
    }
  }
  return fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      ...(userID ? { userid: userID } : {}),
    },
    body: JSON.stringify(body),
    signal,
    mode: 'cors',
  });
}

export interface ParsedSseEvent {
  type: string;
  payload: any;
}

export async function consumeSseStream(
  response: Response,
  onEvent: (event: ParsedSseEvent) => void,
) {
  if (!response.body) {
    throw new Error('missing_response_body');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      return;
    }

    buffer += decoder.decode(value, { stream: true });
    const blocks = buffer.split('\n\n');
    buffer = blocks.pop() || '';

    for (const block of blocks) {
      const eventMatch = block.match(/^event:\s*(.+)$/m);
      const dataMatch = block.match(/^data:\s*(.+)$/m);
      if (!dataMatch) continue;

      const payload = JSON.parse(dataMatch[1]);
      const type = payload?.type || eventMatch?.[1] || 'message';
      onEvent({ type, payload });
    }
  }
}

export function buildSessionSnapshot(input: Partial<ActiveInterviewSession> & Pick<ActiveInterviewSession, 'mode' | 'sessionId' | 'params' | 'conversationHistory'>): ActiveInterviewSession {
  return {
    recordId: input.recordId,
    startedAt: input.startedAt || Date.now(),
    questionIndex: input.questionIndex || 0,
    questionText: input.questionText || '',
    answeredCount: input.answeredCount || 0,
    elapsed: input.elapsed || 0,
    phase: input.phase || 'starting',
    ...input,
  };
}
