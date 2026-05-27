'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { Alert, Button, Descriptions, Skeleton, Tabs, message } from 'antd';
import { BarChartOutlined, ReloadOutlined } from '@ant-design/icons';
import apiClient from '@/services/api/client';

function PrettyBlock({ data }: { data: any }) {
  if (!data) {
    return <div className="text-sm text-[#5f6878]">暂无数据</div>;
  }
  return (
    <pre className="max-h-[560px] overflow-auto rounded-lg bg-[#172033] p-4 text-xs leading-6 text-white">
      {JSON.stringify(data, null, 2)}
    </pre>
  );
}

export default function InterviewResultPage() {
  const params = useParams<{ id: string }>();
  const recordId = Number(params?.id);
  const [loading, setLoading] = useState(false);
  const [report, setReport] = useState<any>(null);
  const [prediction, setPrediction] = useState<any>(null);

  const fetchData = async () => {
    if (!recordId) return;
    setLoading(true);
    try {
      const [reportRes, predictRes] = await Promise.allSettled([
        apiClient.post('/evaluation/report', { record_id: recordId }),
        apiClient.post('/evaluation/predict', { record_id: recordId }),
      ]);

      if (reportRes.status === 'fulfilled') {
        setReport(reportRes.value);
      } else {
        setReport(null);
      }

      if (predictRes.status === 'fulfilled') {
        setPrediction(predictRes.value);
      } else {
        setPrediction(null);
      }

      if (reportRes.status === 'rejected' && predictRes.status === 'rejected') {
        message.error('评估和预测都还没有生成或查询失败');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchData();
  }, [recordId]);

  return (
    <div className="shell py-8">
      <section className="panel rounded-lg p-6">
        <div className="mb-6 flex flex-col justify-between gap-4 md:flex-row md:items-center">
          <div>
            <div className="mb-3 inline-flex items-center gap-2 rounded-lg bg-[#3157b7] px-3 py-2 text-sm font-semibold text-white">
              <BarChartOutlined />
              Evaluation
            </div>
            <h1 className="m-0 text-3xl font-black text-[#172033]">面试结果 #{recordId || '-'}</h1>
            <p className="mt-2 text-sm text-[#5f6878]">
              已对齐后端：POST /evaluation/report 和 POST /evaluation/predict。
            </p>
          </div>
          <Button icon={<ReloadOutlined />} onClick={fetchData} loading={loading}>
            刷新
          </Button>
        </div>

        <Alert
          type="info"
          showIcon
          className="mb-5"
          message="结果由 Kafka 消费者异步生成"
          description="如果刚结束面试后暂时没有数据，等待消费者完成 EvaluationAndPredictService 后再刷新。"
        />

        <Descriptions bordered size="small" column={1} className="mb-5">
          <Descriptions.Item label="record_id">{recordId || '-'}</Descriptions.Item>
          <Descriptions.Item label="报告状态">{report ? '已返回' : '暂无'}</Descriptions.Item>
          <Descriptions.Item label="预测状态">{prediction ? '已返回' : '暂无'}</Descriptions.Item>
        </Descriptions>

        {loading ? (
          <Skeleton active paragraph={{ rows: 8 }} />
        ) : (
          <Tabs
            items={[
              {
                key: 'report',
                label: '评估报告',
                children: <PrettyBlock data={report} />,
              },
              {
                key: 'prediction',
                label: '预测结果',
                children: <PrettyBlock data={prediction} />,
              },
            ]}
          />
        )}
      </section>
    </div>
  );
}
