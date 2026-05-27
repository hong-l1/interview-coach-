'use client';

import Link from 'next/link';
import { Button, Result } from 'antd';

export default function ResetPasswordPage() {
  return (
    <div className="shell py-8">
      <Result
        status="info"
        title="重置密码功能暂未接入"
        subTitle="当前后端路由没有 /user/password/reset，前端已停止调用这个旧接口。"
        extra={
          <Link href="/">
            <Button type="primary" className="ink-button">
              返回工作台
            </Button>
          </Link>
        }
      />
    </div>
  );
}
