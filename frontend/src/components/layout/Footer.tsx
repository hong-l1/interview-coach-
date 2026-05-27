'use client';

import Link from 'next/link';

export default function Footer() {
  return (
    <footer className="border-t border-[rgba(24,32,51,0.1)] bg-white/48">
      <div className="shell flex flex-col gap-3 py-6 text-sm text-[#667085] md:flex-row md:items-center md:justify-between">
        <div>
          <span className="font-black text-[#182033]">面试舱</span>
          <span className="ml-2">AI 面试训练台</span>
        </div>
        <div className="flex flex-wrap gap-4">
          <Link href="/resume" className="hover:text-[#182033]">
            简历库
          </Link>
          <Link href="/user/models" className="hover:text-[#182033]">
            模型配置
          </Link>
          <Link href="/user/interviews" className="hover:text-[#182033]">
            报告
          </Link>
        </div>
      </div>
    </footer>
  );
}
