'use client';

import { useState } from 'react';
import { Button, Card, Descriptions, Tag, message } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, SyncOutlined } from '@ant-design/icons';
import { INTERVIEW_API } from '@/config/api';

type HealthCheckResult = {
  interviewEndpoint: boolean;
  corsSupport: boolean;
  error?: string;
};

export default function BackendHealthCheck() {
  const [checking, setChecking] = useState(false);
  const [result, setResult] = useState<HealthCheckResult | null>(null);

  const checkBackend = async () => {
    setChecking(true);
    const next: HealthCheckResult = {
      interviewEndpoint: false,
      corsSupport: false,
    };

    try {
      const response = await fetch(INTERVIEW_API.START_STREAM, {
        method: 'OPTIONS',
        mode: 'cors',
      });
      next.interviewEndpoint = response.status !== 404;
      next.corsSupport = !!response.headers.get('Access-Control-Allow-Origin');
      setResult(next);
      if (next.interviewEndpoint) {
        message.success('后端面试接口可访问');
      } else {
        message.warning('后端面试接口返回 404');
      }
    } catch (error: any) {
      next.error = error?.message || '网络错误';
      setResult(next);
      message.error('无法连接后端服务');
    } finally {
      setChecking(false);
    }
  };

  const StatusTag = ({ status }: { status: boolean }) =>
    status ? (
      <Tag icon={<CheckCircleOutlined />} color="success">
        正常
      </Tag>
    ) : (
      <Tag icon={<CloseCircleOutlined />} color="error">
        异常
      </Tag>
    );

  return (
    <Card
      title="后端服务诊断"
      extra={
        <Button type="primary" icon={<SyncOutlined />} loading={checking} onClick={checkBackend}>
          检查
        </Button>
      }
    >
      {!result && <p className="text-gray-500">点击检查按钮测试 /mianshi/stream/start。</p>}
      {result && (
        <Descriptions column={1} bordered>
          <Descriptions.Item label="面试接口">
            <StatusTag status={result.interviewEndpoint} />
          </Descriptions.Item>
          <Descriptions.Item label="CORS">
            <StatusTag status={result.corsSupport} />
          </Descriptions.Item>
          {result.error && (
            <Descriptions.Item label="错误">
              <span className="text-red-500">{result.error}</span>
            </Descriptions.Item>
          )}
        </Descriptions>
      )}
    </Card>
  );
}
