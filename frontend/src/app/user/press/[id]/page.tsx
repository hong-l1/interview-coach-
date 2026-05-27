'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { Alert, Button, Skeleton } from 'antd';
import { ReloadOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { predictionService } from '@/services/api/prediction';

export default function PressDetailPage() {
  const params = useParams<{ id: string }>();
  const recordId = Number(params?.id);
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<any>(null);

  const fetchData = async () => {
    if (!recordId) return;
    setLoading(true);
    try {
      setData(await predictionService.getPredictionDetail(recordId));
    } catch {
      setData(null);
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
            <div className="mb-3 inline-flex items-center gap-2 rounded-lg bg-[#d96c4a] px-3 py-2 text-sm font-semibold text-white">
              <ThunderboltOutlined />
              Prediction
            </div>
            <h1 className="m-0 text-3xl font-black text-[#172033]">预测详情 #{recordId || '-'}</h1>
            <p className="mt-2 text-sm text-[#5f6878]">
              旧的 /prediction/:id 已替换为 POST /evaluation/predict。
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
          message="按 record_id 查询预测"
          description="如果这里为空，说明该 record_id 尚未完成评估预测，或 Kafka 消费者还未处理完成。"
        />

        {loading ? (
          <Skeleton active paragraph={{ rows: 8 }} />
        ) : (
          <pre className="max-h-[560px] overflow-auto rounded-lg bg-[#172033] p-4 text-xs leading-6 text-white">
            {JSON.stringify(data || {}, null, 2)}
          </pre>
        )}
      </section>
    </div>
  );
}
