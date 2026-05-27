'use client';

import InterviewStartShell from '@/components/interview/InterviewStartShell';
import { buildSocialRequest } from '@/components/interview/InterviewStartFactories';

export default function SocialInterviewStartPage() {
  return (
    <InterviewStartShell
      mode="social"
      title="社招面试"
      subtitle="社招简历面试"
      buildStartRequest={buildSocialRequest}
    />
  );
}
