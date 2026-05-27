'use client';

import Link from 'next/link';
import { Alert, Button, Empty, Form, InputNumber } from 'antd';
import {
  ArrowRightOutlined,
  BarChartOutlined,
  FileSearchOutlined,
  HistoryOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';

export default function InterviewRecordsPage() {
  const [form] = Form.useForm<{ record_id: number }>();
  const router = useRouter();

  const openResult = async () => {
    const values = await form.validateFields();
    router.push(`/user/interviews/results/${values.record_id}`);
  };

  return (
    <div className="shell py-8">
      <section className="grid gap-5 lg:grid-cols-[1fr_360px]">
        <div className="panel rounded-lg p-8">
          <div className="mb-8 flex flex-col justify-between gap-4 md:flex-row md:items-start">
            <div>
              <div className="mb-3 inline-flex items-center gap-2 rounded-lg bg-[#172033] px-3 py-2 text-sm font-semibold text-white">
                <HistoryOutlined />
                Interview Records
              </div>
              <h1 className="m-0 text-3xl font-black text-[#172033]">面试记录</h1>
              <p className="mt-3 max-w-2xl text-sm leading-6 text-[#5f6878]">
                当前后端还没有开放记录列表接口，结果页已对接评估和预测接口。
              </p>
            </div>
            <Link href="/interview/social">
              <Button type="primary" className="ink-button" icon={<ArrowRightOutlined />}>
                开始面试
              </Button>
            </Link>
          </div>

          <div className="mb-8 rounded-lg border border-[#172033]/10 bg-white/72 p-5">
            <h2 className="m-0 mb-4 text-lg font-black text-[#172033]">按 record_id 查看结果</h2>
            <Form form={form} layout="inline" className="gap-3">
              <Form.Item
                label="record_id"
                name="record_id"
                rules={[{ required: true, message: '请输入 record_id' }]}
              >
                <InputNumber min={1} precision={0} placeholder="例如 1" />
              </Form.Item>
              <Button type="primary" icon={<SearchOutlined />} onClick={openResult} className="ink-button">
                查看结果
              </Button>
            </Form>
          </div>

          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无记录列表数据">
            <div className="flex flex-wrap justify-center gap-3">
              <Link href="/interview/social">
                <Button>社招面试</Button>
              </Link>
              <Link href="/interview/campus">
                <Button>校招面试</Button>
              </Link>
              <Link href="/interview/special">
                <Button>专项面试</Button>
              </Link>
            </div>
          </Empty>
        </div>

        <aside className="space-y-4">
          <div className="surface rounded-lg p-5">
            <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-lg bg-[#3157b7] text-xl text-white">
              <BarChartOutlined />
            </div>
            <div className="text-lg font-black text-[#172033]">评估报告</div>
            <p className="mt-2 text-sm leading-6 text-[#5f6878]">POST /evaluation/report 已接入，用 record_id 查询。</p>
          </div>
          <div className="surface rounded-lg p-5">
            <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-lg bg-[#d96c4a] text-xl text-white">
              <FileSearchOutlined />
            </div>
            <div className="text-lg font-black text-[#172033]">预测结果</div>
            <p className="mt-2 text-sm leading-6 text-[#5f6878]">POST /evaluation/predict 已接入，等待 Kafka 消费后可查。</p>
          </div>
          <Alert
            type="info"
            showIcon
            message="列表接口待补"
            description="完整历史列表需要后端返回 record_id、类型、领域、状态和创建时间。"
          />
        </aside>
      </section>
    </div>
  );
}
