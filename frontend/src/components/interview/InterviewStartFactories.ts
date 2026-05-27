export function getInterviewParams() {
  const inlineParams = (window as any).__interviewParams;
  if (inlineParams) return inlineParams;
  try {
    return JSON.parse(sessionStorage.getItem('interviewParams') || 'null');
  } catch {
    return null;
  }
}

function clean(value: unknown) {
  return String(value ?? '').replace(/[<>&"'`]/g, '').trim();
}

export function buildCampusRequest(params: any) {
  if (!params || !params.resume_id) return null;
  return {
    interview_type: 'comprehensive',
    domain: clean(params.domain || '校招简历面试'),
    difficulty: clean(params.difficulty || '简单'),
    position: clean(params.position || params.position_name || ''),
    company: clean(params.company || params.company_name || ''),
    resume_id: Number(params.resume_id),
  };
}

export function buildSocialRequest(params: any) {
  if (!params || !params.resume_id) return null;
  return {
    interview_type: 'comprehensive',
    domain: clean(params.domain || '社招简历面试'),
    difficulty: clean(params.difficulty || '中等'),
    position: clean(params.position || params.position_name || ''),
    company: clean(params.company || params.company_name || ''),
    resume_id: Number(params.resume_id),
  };
}

export function buildSpecialRequest(params: any) {
  if (!params) return null;
  return {
    interview_type: 'specialized',
    domain: clean(params.domain || ''),
    difficulty: clean(params.difficulty || '中等'),
  };
}
