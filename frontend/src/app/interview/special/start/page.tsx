'use client';

import InterviewStartShell from '@/components/interview/InterviewStartShell';
import { buildSpecialRequest } from '@/components/interview/InterviewStartFactories';

export default function SpecialInterviewStartPage() {
  return (
    <InterviewStartShell
      mode="special"
      title="专项面试"
      subtitle="技能专项训练"
      buildStartRequest={buildSpecialRequest}
    />
  );
}
