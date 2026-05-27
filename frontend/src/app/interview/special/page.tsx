'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button, Form, Select, message } from 'antd';
import {
  ApiOutlined,
  BranchesOutlined,
  DatabaseOutlined,
  PlayCircleOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';

const topicOptions = [
  {
    label: '语言与框架',
    options: [
      { value: 'Go', label: 'Go' },
      { value: 'Java', label: 'Java' },
      { value: 'Node.js', label: 'Node.js' },
      { value: 'React', label: 'React' },
    ],
  },
  {
    label: '后端组件',
    options: [
      { value: 'MySQL', label: 'MySQL' },
      { value: 'Redis', label: 'Redis' },
      { value: 'Kafka', label: 'Kafka' },
      { value: '微服务', label: '微服务' },
    ],
  },
  {
    label: '计算机基础',
    options: [
      { value: '操作系统', label: '操作系统' },
      { value: '计算机网络', label: '计算机网络' },
      { value: '数据结构与算法', label: '数据结构与算法' },
    ],
  },
];

const difficultyOptions = [
  { value: '简单', label: '简单' },
  { value: '中等', label: '中等' },
  { value: '困难', label: '困难' },
];

export default function SpecialInterviewPage() {
  const [form] = Form.useForm();
  const [topic, setTopic] = useState('Go');
  const [starting, setStarting] = useState(false);
  const router = useRouter();

  const handleStart = async () => {
    try {
      const values = await form.validateFields();
      const params = {
        domain: values.domain,
        difficulty: values.difficulty,
      };
      (window as any).__interviewParams = params;
      window.sessionStorage.setItem('interviewParams', JSON.stringify(params));
      setStarting(true);
      router.push('/interview/special/start');
    } catch {
      message.warning('请选择专项方向和难度');
    }
  };

  return (
    <div className="shell py-8">
      <section className="grid gap-5 lg:grid-cols-[1fr_420px]">
        <div className="panel rounded-lg p-8">
          <div className="mb-10 max-w-3xl">
            <div className="mb-3 inline-flex items-center gap-2 rounded-lg bg-[#d96c4a] px-3 py-2 text-sm font-semibold text-white">
              <ThunderboltOutlined />
              专项突破
            </div>
            <h1 className="m-0 text-4xl font-black leading-tight text-[#182033]">
              选择一个技术点，让面试官只围绕它持续追问。
            </h1>
            <p className="mt-4 text-base leading-7 text-[#667085]">
              专项面试不依赖简历，更适合查漏补缺和单点强化。后端会调用 specialized 类型的面试 agent。
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-3">
            {[
              { icon: <ApiOutlined />, title: '单点聚焦', desc: '围绕一个技术域持续追问，不分散。' },
              { icon: <BranchesOutlined />, title: '链路展开', desc: '从概念、原理、实现到场景逐层深入。' },
              { icon: <DatabaseOutlined />, title: '暴露盲区', desc: '快速定位知识断点和表达卡点。' },
            ].map((item) => (
              <div key={item.title} className="surface rounded-lg p-5">
                <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg bg-[#182033] text-lg text-white">
                  {item.icon}
                </div>
                <div className="font-bold text-[#182033]">{item.title}</div>
                <p className="m-0 mt-2 text-sm leading-6 text-[#667085]">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="panel rounded-lg bg-white p-6">
          <h2 className="m-0 mb-1 text-xl font-black text-[#182033]">专项配置</h2>
          <p className="mb-6 text-sm text-[#667085]">当前方向：{topic}</p>
          <Form form={form} layout="vertical" initialValues={{ domain: topic, difficulty: '中等' }}>
            <Form.Item label="专项方向" name="domain" rules={[{ required: true }]}>
              <Select options={topicOptions} onChange={setTopic} />
            </Form.Item>
            <Form.Item label="难度" name="difficulty" rules={[{ required: true }]}>
              <Select options={difficultyOptions} />
            </Form.Item>
            <Button
              type="primary"
              block
              size="large"
              icon={<PlayCircleOutlined />}
              loading={starting}
              onClick={handleStart}
              className="ink-button !h-12"
            >
              开始专项面试
            </Button>
          </Form>
        </div>
      </section>
    </div>
  );
}
