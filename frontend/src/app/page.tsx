'use client';

import Link from 'next/link';
import { Button } from 'antd';
import {
  ArrowRightOutlined,
  AuditOutlined,
  BarChartOutlined,
  BranchesOutlined,
  DatabaseOutlined,
  FileSearchOutlined,
  MessageOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';

const stats = [
  ['20', '轮追问上限'],
  ['Redis', '断线恢复'],
  ['Kafka', '异步评估'],
  ['RAG', '专项检索'],
];

const modes = [
  {
    title: '社招综合面试',
    desc: '围绕简历、目标公司、岗位和难度持续追问，适合模拟完整后端面试。',
    href: '/interview/social',
    icon: <MessageOutlined />,
    tone: '#d96c4a',
  },
  {
    title: '校招综合面试',
    desc: '从基础、项目和表达能力三条线评估，更适合应届生和实习场景。',
    href: '/interview/campus',
    icon: <FileSearchOutlined />,
    tone: '#3157b7',
  },
  {
    title: '专项拷问',
    desc: '聚焦语言、数据库、中间件、网络与操作系统，配合知识库混合检索。',
    href: '/interview/special',
    icon: <DatabaseOutlined />,
    tone: '#57765a',
  },
];

const flow = [
  { title: '上传简历', desc: '解析候选人背景和项目线索，综合面试使用选中的简历作为上下文。', icon: <FileSearchOutlined /> },
  { title: '开始面试', desc: 'SSE 持续推送问题，答案按轮提交到后端运行时。', icon: <BranchesOutlined /> },
  { title: '生成报告', desc: '面试结束后发 Kafka 消息，消费者生成评估和预测。', icon: <AuditOutlined /> },
];

export default function Home() {
  return (
    <div>
      <section className="shell grid min-h-[calc(100vh-64px)] items-center gap-10 py-12 lg:grid-cols-[1.05fr_0.95fr]">
        <div className="animate-rise">
          <div className="mb-5 inline-flex items-center gap-2 rounded-lg border border-[rgba(24,32,51,0.1)] bg-white/70 px-3 py-2 text-sm font-semibold text-[#182033]/70">
            <ThunderboltOutlined className="text-[#d96c4a]" />
            AI Interview Operating System
          </div>
          <h1 className="max-w-3xl text-5xl font-black leading-[1.04] text-[#182033] md:text-7xl">
            把面试训练变成可复盘的系统工程。
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-8 text-[#667085]">
            从简历解析、实时追问、断线恢复，到结束后触发 Kafka 评估与预测题生成。
            这里不是展示页，是一套围绕真实后端链路运行的面试训练台。
          </p>
          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Link href="/interview/social">
              <Button className="ink-button h-12 px-6" icon={<ArrowRightOutlined />}>
                开始综合面试
              </Button>
            </Link>
            <Link href="/resume">
              <Button className="h-12 border-[rgba(24,32,51,0.14)] bg-white/70 px-6 font-semibold text-[#182033]">
                管理简历库
              </Button>
            </Link>
          </div>
          <div className="mt-10 grid grid-cols-2 gap-3 sm:grid-cols-4">
            {stats.map(([value, label]) => (
              <div key={label} className="surface rounded-lg p-4">
                <div className="text-2xl font-black text-[#182033]">{value}</div>
                <div className="mt-1 text-xs font-semibold text-[#667085]">{label}</div>
              </div>
            ))}
          </div>
        </div>

        <div className="panel animate-rise rounded-lg p-4 md:p-6" style={{ animationDelay: '90ms' }}>
          <div className="rounded-lg bg-[#182033] p-5 text-white">
            <div className="flex items-center justify-between border-b border-white/10 pb-4">
              <div>
                <div className="text-sm text-white/50">Live Interview</div>
                <div className="mt-1 text-xl font-black">Go Runtime 深挖</div>
              </div>
              <div className="rounded-lg bg-[#d96c4a] px-3 py-1 text-xs font-black">Round 07</div>
            </div>
            <div className="space-y-4 py-6">
              <div className="max-w-[92%] rounded-lg bg-white/10 p-4 leading-7">
                你说 goroutine 调度很轻量，那 G-M-P 模型里 P 的本质作用是什么？
              </div>
              <div className="ml-auto max-w-[86%] rounded-lg bg-[#fffaf1] p-4 leading-7 text-[#182033]">
                P 维护本地运行队列，持有执行 Go 代码所需的调度资源，同时减少全局队列竞争。
              </div>
              <div className="max-w-[92%] rounded-lg bg-white/10 p-4 leading-7">
                如果本地队列满了，新建的 G 会怎么处理？再解释 work stealing 的触发点。
              </div>
            </div>
            <div className="grid grid-cols-3 gap-3 border-t border-white/10 pt-4 text-center">
              <div>
                <div className="text-2xl font-black">82</div>
                <div className="text-xs text-white/45">表达</div>
              </div>
              <div>
                <div className="text-2xl font-black">74</div>
                <div className="text-xs text-white/45">深度</div>
              </div>
              <div>
                <div className="text-2xl font-black">5</div>
                <div className="text-xs text-white/45">薄弱点</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="shell py-14">
        <div className="mb-8 flex items-end justify-between gap-6">
          <div>
            <p className="text-sm font-black uppercase tracking-[0.24em] text-[#d96c4a]">Modes</p>
            <h2 className="mt-2 text-3xl font-black text-[#182033] md:text-4xl">三种训练工作流</h2>
          </div>
          <Link href="/user/interviews" className="hidden text-sm font-bold text-[#3157b7] md:block">
            查看报告 <ArrowRightOutlined />
          </Link>
        </div>
        <div className="grid gap-4 md:grid-cols-3">
          {modes.map((item) => (
            <Link
              key={item.title}
              href={item.href}
              className="surface group rounded-lg p-6 transition hover:-translate-y-1 hover:bg-white"
            >
              <div className="mb-8 grid h-12 w-12 place-items-center rounded-lg text-xl text-white" style={{ background: item.tone }}>
                {item.icon}
              </div>
              <h3 className="text-xl font-black text-[#182033]">{item.title}</h3>
              <p className="mt-3 min-h-20 leading-7 text-[#667085]">{item.desc}</p>
              <div className="mt-6 text-sm font-black text-[#3157b7] transition group-hover:translate-x-1">
                进入模块 <ArrowRightOutlined />
              </div>
            </Link>
          ))}
        </div>
      </section>

      <section className="shell py-14">
        <div className="panel rounded-lg p-6 md:p-8">
          <div className="mb-6 flex items-center gap-3">
            <BarChartOutlined className="text-2xl text-[#d96c4a]" />
            <h2 className="m-0 text-2xl font-black text-[#182033]">真实链路</h2>
          </div>
          <div className="grid gap-4 md:grid-cols-3">
            {flow.map((item, index) => (
              <div key={item.title} className="rounded-lg bg-white/70 p-5">
                <div className="mb-5 flex items-center gap-3">
                  <div className="grid h-10 w-10 place-items-center rounded-lg bg-[#57765a] text-white">
                    {item.icon}
                  </div>
                  <div className="text-sm font-black text-[#182033]/35">0{index + 1}</div>
                </div>
                <h3 className="text-lg font-black text-[#182033]">{item.title}</h3>
                <p className="mt-2 leading-7 text-[#667085]">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
