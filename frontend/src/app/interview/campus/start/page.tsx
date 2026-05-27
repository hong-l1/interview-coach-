'use client';

import InterviewStartShell from '@/components/interview/InterviewStartShell';
import { buildCampusRequest } from '@/components/interview/InterviewStartFactories';

export default function CampusInterviewStartPage() {
  return (
    <InterviewStartShell
      mode="campus"
      title="校招面试"
      subtitle="校招简历面试"
      buildStartRequest={buildCampusRequest}
    />
  );
}
