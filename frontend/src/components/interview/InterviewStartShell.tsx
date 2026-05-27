'use client';

import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Avatar, Button, Input, Progress, message } from 'antd';
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  LoadingOutlined,
  LogoutOutlined,
  SendOutlined,
  UserOutlined,
} from '@ant-design/icons';

import { INTERVIEW_API } from '@/config/api';
import {
  ActiveInterviewSession,
  ConversationItem,
  InterviewMode,
  buildSessionSnapshot,
  buildSseRequest,
  clearActiveInterviewSession,
  consumeSseStream,
  fetchSessionInfo,
  getAuthToken,
  loadActiveInterviewSession,
  saveActiveInterviewSession,
} from '@/utils/interviewSession';

type Props = {
  mode: InterviewMode;
  title: string;
  subtitle: string;
  accentClassName?: string;
  secondaryAccentClassName?: string;
  buildStartRequest: (params: any) => Record<string, any> | null;
};

function unwrapApiResponse(payload: any) {
  if (payload && typeof payload === 'object' && 'code' in payload) {
    return payload.data;
  }
  return payload;
}

export default function InterviewStartShell({ mode, title, subtitle, buildStartRequest }: Props) {
  const router = useRouter();
  const [elapsed, setElapsed] = useState(0);
  const [sessionId, setSessionId] = useState('');
  const [questionIndex, setQuestionIndex] = useState(0);
  const [answer, setAnswer] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [streaming, setStreaming] = useState(true);
  const [waitingNextQuestion, setWaitingNextQuestion] = useState(false);
  const [reconnecting, setReconnecting] = useState(false);
  const [conversationHistory, setConversationHistory] = useState<ConversationItem[]>([]);
  const [pageSubtitle, setPageSubtitle] = useState(subtitle);

  const abortRef = useRef<AbortController | null>(null);
  const sessionRef = useRef<ActiveInterviewSession | null>(null);
  const paramsRef = useRef<Record<string, any>>({});
  const chatRef = useRef<HTMLDivElement>(null);
  const unmountedRef = useRef(false);

  useEffect(() => {
    const timer = window.setInterval(() => setElapsed((prev) => prev + 1), 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    chatRef.current?.scrollTo({ top: chatRef.current.scrollHeight, behavior: 'smooth' });
  }, [conversationHistory, waitingNextQuestion, streaming, reconnecting]);

  useEffect(() => {
    if (sessionRef.current) {
      persistSession({ elapsed });
    }
  }, [elapsed]);

  useEffect(() => {
    let params = (window as any).__interviewParams;
    if (!params) {
      try {
        params = JSON.parse(window.sessionStorage.getItem('interviewParams') || 'null');
      } catch {
        params = null;
      }
    }

    const requestBody = buildStartRequest(params);
    if (!requestBody) {
      message.error('缺少面试参数，请从面试配置页重新进入');
      setStreaming(false);
      return;
    }

    paramsRef.current = params || {};
    if (params?.domain) setPageSubtitle(String(params.domain));

    const boot = async () => {
      const token = getAuthToken();
      if (!token) {
        message.error('请先登录后再开始面试');
        setStreaming(false);
        return;
      }

      const saved = loadActiveInterviewSession(mode);
      if (saved?.sessionId && (await restoreSession(saved, token))) {
        return;
      }

      await openStream(token, requestBody, false);
    };

    void boot();

    return () => {
      unmountedRef.current = true;
      abortRef.current?.abort();
    };
  }, [buildStartRequest, mode]);

  const persistSession = (patch: Partial<ActiveInterviewSession>) => {
    if (!sessionRef.current) return;
    sessionRef.current = { ...sessionRef.current, ...patch };
    saveActiveInterviewSession(mode, sessionRef.current);
  };

  const rememberSession = (next: ActiveInterviewSession) => {
    sessionRef.current = next;
    saveActiveInterviewSession(mode, next);
  };

  const restoreSession = async (saved: ActiveInterviewSession, token: string) => {
    try {
      const raw = await fetchSessionInfo(saved.sessionId, token);
      const data = unwrapApiResponse(raw);
      const status = data?.status || data?.session?.status;
      if (status !== 'active') {
        clearActiveInterviewSession(mode);
        return false;
      }

      const restored = buildSessionSnapshot({
        ...saved,
        elapsed: saved.elapsed || Number(data?.elapsed_time || 0),
        questionIndex: saved.questionIndex || Number(data?.current_index || 0),
        conversationHistory: saved.conversationHistory || [],
        phase: 'resuming',
      });
      setSessionId(restored.sessionId);
      setQuestionIndex(restored.questionIndex);
      setElapsed(restored.elapsed);
      setConversationHistory(restored.conversationHistory);
      rememberSession(restored);
      await openStream(token, { session_id: restored.sessionId }, true);
      return true;
    } catch {
      clearActiveInterviewSession(mode);
      return false;
    }
  };

  const openStream = async (token: string, body: Record<string, any>, resume: boolean) => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setStreaming(true);
    setReconnecting(resume);

    try {
      const response = await buildSseRequest(token, body, resume, controller.signal);
      if (!response.ok) throw new Error(`stream_${response.status}`);

      await consumeSseStream(response, ({ type, payload }) => {
        handleSseEvent(type, payload, resume);
      });
    } catch (error: any) {
      if (error?.name === 'AbortError' || unmountedRef.current) return;
      message.error(resume ? '恢复面试失败，请重新进入' : '面试连接失败，请稍后重试');
      setStreaming(false);
      setReconnecting(false);
    }
  };

  const handleSseEvent = (type: string, payload: any, isResume: boolean) => {
    if (type === 'interview_begin' || type === 'interview_reconnected') {
      const sid = payload?.session_id || '';
      if (!sid) return;
      setSessionId(sid);
      rememberSession(
        buildSessionSnapshot({
          mode,
          sessionId: sid,
          params: paramsRef.current,
          startedAt: Date.now(),
          questionIndex,
          questionText: '',
          answeredCount: conversationHistory.filter((item) => item.type === 'answer').length,
          elapsed,
          conversationHistory,
          phase: isResume ? 'resuming' : 'starting',
        }),
      );
      return;
    }

    if (type === 'question') {
      const q = String(payload?.question || payload?.data?.question || '');
      const idx = Number(payload?.questionIndex ?? payload?.index ?? 0);
      if (!q) return;
      setQuestionIndex(idx);
      setWaitingNextQuestion(false);
      setStreaming(false);
      setReconnecting(false);
      setConversationHistory((prev) => {
        const exists = prev.some((item) => item.type === 'question' && item.index === idx && item.content === q);
        const next = exists ? prev : [...prev, { type: 'question' as const, content: q, index: idx, timestamp: Date.now() }];
        persistSession({
          questionIndex: idx,
          questionText: q,
          conversationHistory: next,
          phase: 'waiting_answer',
        });
        return next;
      });
      return;
    }

    if (type === 'interview_end') {
      setWaitingNextQuestion(false);
      setStreaming(false);
      setReconnecting(false);
      persistSession({ phase: 'completed' });
      clearActiveInterviewSession(mode);
      message.success('面试已完成，评估会在后台生成');
      return;
    }

    if (type === 'error') {
      setWaitingNextQuestion(false);
      setStreaming(false);
      setReconnecting(false);
      message.error(payload?.message || '面试过程出现异常');
    }
  };

  const endInterview = async () => {
    if (!sessionId) {
      router.push('/user/interviews');
      return;
    }
    try {
      const token = getAuthToken();
      let userID = '';
      try {
        const user = JSON.parse(window.localStorage.getItem('user') || 'null');
        userID = String(user?.id || user?.user_id || user?.ID || '');
      } catch {
        userID = '';
      }
      await fetch(INTERVIEW_API.END_INTERVIEW, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
          ...(userID ? { userid: userID } : {}),
        },
        body: JSON.stringify({ session_id: sessionId }),
      });
    } catch {
      // The local session still needs to be cleared when the user leaves.
    }
    abortRef.current?.abort();
    clearActiveInterviewSession(mode);
    router.push('/user/interviews');
  };

  const submitAnswer = async () => {
    const currentAnswer = answer.trim();
    if (!sessionId) {
      message.warning('会话还没有建立，请稍等');
      return;
    }
    if (!currentAnswer) {
      message.warning('请输入回答后再提交');
      return;
    }

    const token = getAuthToken();
    if (!token) {
      message.error('登录已过期，请重新登录');
      return;
    }

    setSubmitting(true);
    setWaitingNextQuestion(true);
    setAnswer('');
    setConversationHistory((prev) => {
      const next = [...prev, { type: 'answer' as const, content: currentAnswer, timestamp: Date.now() }];
      persistSession({ conversationHistory: next, phase: 'submitting_answer' });
      return next;
    });

    try {
      let userID = '';
      try {
        const user = JSON.parse(window.localStorage.getItem('user') || 'null');
        userID = String(user?.id || user?.user_id || user?.ID || '');
      } catch {
        userID = '';
      }
      const response = await fetch(INTERVIEW_API.SUBMIT_ANSWER, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
          ...(userID ? { userid: userID } : {}),
        },
        body: JSON.stringify({ session_id: sessionId, answer: currentAnswer }),
      });
      if (!response.ok) throw new Error(`submit_${response.status}`);
      persistSession({
        answeredCount: conversationHistory.filter((item) => item.type === 'answer').length + 1,
        phase: 'answer_submitted',
      });
    } catch {
      setWaitingNextQuestion(false);
      message.error('提交失败，请稍后重试');
    } finally {
      setSubmitting(false);
    }
  };

  const minutes = String(Math.floor(elapsed / 60)).padStart(2, '0');
  const seconds = String(elapsed % 60).padStart(2, '0');
  const answerCount = conversationHistory.filter((item) => item.type === 'answer').length;
  const percent = Math.min(100, Math.round((Math.max(questionIndex, answerCount) / 20) * 100));
  const phaseText = reconnecting ? '恢复中' : waitingNextQuestion ? '生成中' : streaming ? '连接中' : '答题中';
  const canType = !submitting && !waitingNextQuestion && !streaming && !reconnecting;

  return (
    <div className="min-h-screen bg-[#f6f1e8]">
      <header className="border-b border-[#172033]/10 bg-[#fffaf1]/88 backdrop-blur-xl">
        <div className="shell flex min-h-20 flex-col justify-center gap-4 py-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="min-w-0">
            <div className="mb-2 flex flex-wrap items-center gap-2">
              <span className="rounded-full bg-[#d96c4a]/12 px-3 py-1 text-xs font-black uppercase tracking-[0.22em] text-[#b34f31]">
                Live Interview
              </span>
              <span className="rounded-full border border-[#172033]/10 bg-white px-3 py-1 text-xs font-bold text-[#172033]">
                {phaseText}
              </span>
            </div>
            <h1 className="m-0 truncate text-2xl font-black text-[#172033]">{title}</h1>
            <p className="m-0 mt-1 text-sm text-[#5f6878]">{pageSubtitle}</p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <div className="surface flex min-w-[116px] items-center gap-2 rounded-lg px-3 py-2 text-sm text-[#172033]">
              <ClockCircleOutlined />
              <span className="font-mono">
                {minutes}:{seconds}
              </span>
            </div>
            <div className="surface hidden min-w-[116px] rounded-lg px-3 py-2 text-sm text-[#172033] sm:block">
              <span className="font-bold">{answerCount}</span>
              <span className="ml-1 text-[#5f6878]">/ 20 已答</span>
            </div>
            <Button icon={<LogoutOutlined />} onClick={endInterview} className="!h-10">
              结束
            </Button>
          </div>
        </div>
      </header>

      <main className="shell grid min-h-[calc(100vh-96px)] grid-cols-1 gap-5 py-5 lg:grid-cols-[300px_1fr]">
        <aside className="space-y-4">
          <section className="panel rounded-lg p-5">
            <div className="mb-4 flex items-end justify-between gap-3">
              <div>
                <div className="text-xs font-black uppercase tracking-[0.18em] text-[#d96c4a]">Progress</div>
                <div className="mt-1 text-3xl font-black text-[#172033]">{percent}%</div>
              </div>
              <div className="rounded-lg bg-[#172033] px-3 py-2 text-right text-white">
                <div className="font-mono text-lg leading-none">{questionIndex || '--'}</div>
                <div className="mt-1 text-[11px] text-white/70">当前题</div>
              </div>
            </div>
            <Progress percent={percent} showInfo={false} strokeColor="#d96c4a" />
            <div className="mt-5 grid grid-cols-2 gap-3">
              <div className="rounded-lg border border-[#172033]/10 bg-white/78 p-3">
                <div className="text-2xl font-black text-[#172033]">{answerCount}</div>
                <div className="mt-1 text-xs text-[#5f6878]">已回答</div>
              </div>
              <div className="rounded-lg border border-[#172033]/10 bg-white/78 p-3">
                <div className="text-2xl font-black text-[#172033]">{20 - Math.min(answerCount, 20)}</div>
                <div className="mt-1 text-xs text-[#5f6878]">剩余题</div>
              </div>
            </div>
          </section>

          <section className="surface rounded-lg p-5">
            <div className="mb-4 text-sm font-black text-[#172033]">会话状态</div>
            <div className="space-y-3 text-sm">
              <div className="flex items-center justify-between gap-3">
                <span className="text-[#5f6878]">Session</span>
                <span className="max-w-[150px] truncate font-mono text-xs text-[#172033]">{sessionId || '等待建立'}</span>
              </div>
              <div className="flex items-center justify-between gap-3">
                <span className="text-[#5f6878]">连接</span>
                <span className="font-bold text-[#5d7d5a]">{phaseText}</span>
              </div>
              <div className="flex items-center justify-between gap-3">
                <span className="text-[#5f6878]">保存</span>
                <span className="font-bold text-[#3157b7]">实时同步</span>
              </div>
            </div>
          </section>
        </aside>

        <section className="panel flex min-h-[74vh] overflow-hidden rounded-lg">
          <div className="flex min-w-0 flex-1 flex-col">
            <div className="flex flex-col gap-3 border-b border-[#172033]/10 bg-white/54 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <div className="text-sm font-black text-[#172033]">面试对话</div>
                <div className="mt-1 text-xs text-[#5f6878]">第 {questionIndex || '-'} 题 · {answerCount} 次回答</div>
              </div>
              <div className="flex flex-wrap gap-2 text-xs font-bold">
                <span className="rounded-full bg-[#3157b7]/10 px-3 py-1 text-[#3157b7]">题目</span>
                <span className="rounded-full bg-[#5d7d5a]/12 px-3 py-1 text-[#4c704f]">回答</span>
              </div>
            </div>

            <div ref={chatRef} className="flex-1 overflow-y-auto bg-[#fffaf1]/48 p-4 sm:p-6">
            {conversationHistory.length === 0 && (
              <div className="flex h-full min-h-[360px] flex-col items-center justify-center text-center">
                <div className="mb-5 flex h-16 w-16 items-center justify-center rounded-lg bg-[#172033] text-2xl text-white shadow-lg shadow-[#172033]/15">
                  {streaming ? <LoadingOutlined /> : <CheckCircleOutlined />}
                </div>
                <h2 className="m-0 text-2xl font-black text-[#172033]">正在准备第一题</h2>
                <p className="mt-2 max-w-md text-sm leading-6 text-[#5f6878]">连接建立后，第一道题会自动出现在这里。</p>
              </div>
            )}

            <div className="space-y-5">
              {conversationHistory.map((item, index) => {
                const isQuestion = item.type === 'question';
                return (
                  <div key={`${item.type}-${index}-${item.timestamp}`} className={`flex gap-3 ${isQuestion ? '' : 'flex-row-reverse'}`}>
                    <Avatar
                      icon={isQuestion ? <CheckCircleOutlined /> : <UserOutlined />}
                      className={isQuestion ? '!bg-[#3157b7]' : '!bg-[#5d7d5a]'}
                    />
                    <div className={`max-w-[84%] rounded-lg px-4 py-3 shadow-sm ${isQuestion ? 'border border-[#3157b7]/12 bg-white text-[#172033]' : 'bg-[#172033] text-white'}`}>
                      {isQuestion && (
                        <div className="mb-1 text-xs font-bold uppercase tracking-[0.18em] text-[#3157b7]">
                          Question {item.index || index + 1}
                        </div>
                      )}
                      <div className="whitespace-pre-wrap break-words text-sm leading-7">{item.content}</div>
                    </div>
                  </div>
                );
              })}

              {(waitingNextQuestion || streaming || reconnecting) && conversationHistory.length > 0 && (
                <div className="flex items-center gap-3 text-sm text-[#5f6878]">
                  <Avatar icon={<LoadingOutlined />} className="!bg-[#d96c4a]" />
                  <div className="rounded-lg bg-white px-4 py-3">
                    {reconnecting ? '正在恢复连接...' : waitingNextQuestion ? '正在生成下一题...' : '正在等待题目...'}
                  </div>
                </div>
              )}
            </div>
          </div>

            <div className="border-t border-[#172033]/10 bg-white/82 p-4">
              <div className="mb-3 flex items-center justify-between text-xs text-[#5f6878]">
                <span className="font-bold text-[#172033]">本轮回答</span>
                <span>{answer.length} 字</span>
              </div>
              <div className="flex flex-col gap-3 sm:flex-row">
              <Input.TextArea
                value={answer}
                onChange={(event) => setAnswer(event.target.value)}
                autoSize={{ minRows: 2, maxRows: 5 }}
                disabled={!canType}
                placeholder={waitingNextQuestion ? '等待下一题生成中...' : '输入本轮回答，点击按钮提交'}
              />
              <Button
                type="primary"
                icon={<SendOutlined />}
                loading={submitting}
                disabled={!answer.trim() || submitting || waitingNextQuestion || streaming || reconnecting}
                onClick={submitAnswer}
                className="ink-button min-h-12 sm:w-32"
              >
                提交
              </Button>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
