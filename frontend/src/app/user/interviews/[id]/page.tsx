'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Button, Result } from 'antd';

export default function InterviewDetailRedirectPage() {
  const params = useParams<{ id: string }>();
  const id = params?.id;

  return (
    <div className="shell py-8">
      <Result
        status="info"
        title="面试详情页已合并到结果页"
        subTitle="当前后端没有单独的面试详情接口，请通过 record_id 查看评估报告和预测结果。"
        extra={
          <Link href={`/user/interviews/results/${id}`}>
            <Button type="primary" className="ink-button">
              查看结果
            </Button>
          </Link>
        }
      />
    </div>
  );
}
