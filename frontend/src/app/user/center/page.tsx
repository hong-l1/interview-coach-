'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Button, Statistic } from 'antd';
import {
  ArrowRightOutlined,
  CloudServerOutlined,
  FileTextOutlined,
  HistoryOutlined,
  MessageOutlined,
  ThunderboltOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { resumeService } from '@/services/api/resume';

type LocalUser = {
  id?: number;
  username?: string;
  email?: string;
};

export default function UserCenterPage() {
  const [user, setUser] = useState<LocalUser | null>(null);
  const [resumeCount, setResumeCount] = useState(0);

  useEffect(() => {
    try {
      setUser(JSON.parse(window.localStorage.getItem('user') || 'null'));
    } catch {
      setUser(null);
    }

    resumeService
      .list()
      .then((data) => setResumeCount(data.total))
      .catch(() => setResumeCount(0));
  }, []);

  return (
    <div className="shell py-8">
      <section className="grid gap-5 lg:grid-cols-[360px_1fr]">
        <aside className="panel h-fit overflow-hidden rounded-lg">
          <div className="bg-[#172033] p-6 text-white">
            <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-lg bg-white/12 text-2xl">
              <UserOutlined />
            </div>
            <h1 className="m-0 text-2xl font-black">{user?.username || '个人中心'}</h1>
            <p className="mt-2 text-sm text-white/62">{user?.email || '登录后可查看账号信息'}</p>
            <div className="mt-5 grid grid-cols-2 gap-3">
              <div className="rounded-lg bg-white/10 p-3">
                <div className="text-xl font-black">{user?.id ?? '-'}</div>
                <div className="mt-1 text-xs text-white/55">用户 ID</div>
              </div>
              <div className="rounded-lg bg-white/10 p-3">
                <div className="text-xl font-black">{resumeCount}</div>
                <div className="mt-1 text-xs text-white/55">简历数</div>
              </div>
            </div>
          </div>
          <div className="p-5">
            <Link href="/resume">
              <Button type="primary" block className="ink-button h-11" icon={<FileTextOutlined />}>
                管理简历库
              </Button>
            </Link>
            <Link href="/interview/social">
              <Button block className="mt-3 h-11" icon={<MessageOutlined />}>
                开始综合面试
              </Button>
            </Link>
          </div>
        </aside>

        <main className="space-y-5">
          <section className="grid gap-5 md:grid-cols-2">
            <div className="surface rounded-lg p-6">
              <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-lg bg-[#3157b7] text-xl text-white">
                <FileTextOutlined />
              </div>
              <Statistic title="已上传简历" value={resumeCount} suffix="份" />
              <Link href="/resume">
                <Button className="mt-5" type="primary" icon={<ArrowRightOutlined />}>
                  进入简历库
                </Button>
              </Link>
            </div>

            <div className="surface rounded-lg p-6">
              <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-lg bg-[#d96c4a] text-xl text-white">
                <CloudServerOutlined />
              </div>
              <Statistic title="模型配置" value="OpenAI" />
              <Link href="/user/models">
                <Button className="mt-5" icon={<ArrowRightOutlined />}>
                  配置模型
                </Button>
              </Link>
            </div>
          </section>

          <section className="panel rounded-lg p-5">
            <div className="mb-4 flex items-center justify-between gap-3">
              <div>
                <div className="text-sm font-black uppercase tracking-[0.18em] text-[#d96c4a]">Workspace</div>
                <h2 className="m-0 mt-1 text-2xl font-black text-[#172033]">常用入口</h2>
              </div>
              <Link href="/user/interviews" className="text-sm font-bold text-[#3157b7]">
                面试结果 <ArrowRightOutlined />
              </Link>
            </div>
            <div className="grid gap-3 md:grid-cols-3">
              {[
                { href: '/interview/social', title: '社招面试', desc: '按简历和岗位追问', icon: <MessageOutlined />, tone: '#3157b7' },
                { href: '/interview/special', title: '专项训练', desc: '聚焦技术薄弱点', icon: <ThunderboltOutlined />, tone: '#d96c4a' },
                { href: '/user/interviews', title: '结果查询', desc: '按 record_id 查看报告', icon: <HistoryOutlined />, tone: '#57765a' },
              ].map((item) => (
                <Link key={item.href} href={item.href} className="rounded-lg border border-[#172033]/10 bg-white/70 p-4 transition hover:-translate-y-1 hover:bg-white">
                  <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg text-lg text-white" style={{ backgroundColor: item.tone }}>
                    {item.icon}
                  </div>
                  <div className="font-black text-[#172033]">{item.title}</div>
                  <div className="mt-1 text-sm text-[#5f6878]">{item.desc}</div>
                </Link>
              ))}
            </div>
          </section>
        </main>
      </section>
    </div>
  );
}
