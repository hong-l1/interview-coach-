'use client';

import Link from 'next/link';
import { Alert, Button, Empty } from 'antd';
import { FileTextOutlined, ThunderboltOutlined } from '@ant-design/icons';

export default function PressRecordsPage() {
  return (
    <div className="shell py-8">
      <section className="panel rounded-lg p-8">
        <div className="mb-8">
          <div className="mb-3 inline-flex items-center gap-2 rounded-lg bg-[#d96c4a] px-3 py-2 text-sm font-semibold text-white">
            <ThunderboltOutlined />
            押题记录
          </div>
          <h1 className="m-0 text-3xl font-black text-[#172033]">押题功能等待后端路由接入</h1>
          <p className="mt-3 max-w-2xl text-sm leading-6 text-[#5f6878]">
            当前后端没有 /prediction/list、/prediction/:id、/prediction/start。前端已停止调用这些旧接口，
            避免页面加载时产生无效请求。
          </p>
        </div>
        <Alert
          type="info"
          showIcon
          message="当前可用的是面试后的评估与预测"
          description="后端已提供 POST /evaluation/report 和 POST /evaluation/predict，可在面试结果页按 record_id 查询。"
          className="mb-6"
        />
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="暂无押题记录接口"
        >
          <Link href="/interview/special">
            <Button type="primary" icon={<FileTextOutlined />} className="ink-button">
              去专项面试
            </Button>
          </Link>
        </Empty>
      </section>
    </div>
  );
}
