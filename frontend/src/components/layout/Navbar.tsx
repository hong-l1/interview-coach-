'use client';

import { Button, Dropdown, Form, Input, Modal, message } from 'antd';
import {
  LoginOutlined,
  LogoutOutlined,
  MenuOutlined,
  RocketOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons';
import Link from 'next/link';
import { useEffect, useState } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import apiClient from '@/services/api/client';

type UserInfo = {
  id?: number;
  username?: string;
  email?: string;
};

const navItems = [
  { href: '/', label: '工作台' },
  { href: '/resume', label: '简历库' },
  { href: '/interview/social', label: '社招' },
  { href: '/interview/campus', label: '校招' },
  { href: '/interview/special', label: '专项' },
  { href: '/user/interviews', label: '报告' },
];

export default function Navbar() {
  const router = useRouter();
  const pathname = usePathname();
  const [authOpen, setAuthOpen] = useState(false);
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [user, setUser] = useState<UserInfo | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    try {
      const raw = window.localStorage.getItem('user');
      const token = window.localStorage.getItem('token');
      if (token && raw) setUser(JSON.parse(raw));
    } catch {
      setUser(null);
    }
  }, []);

  const submitAuth = async (values: any) => {
    try {
      const url = mode === 'login' ? '/user/login' : '/user/register';
      const data: any = await apiClient.post(url, values);
      const token = data?.token;
      if (!token) {
        message.error('后端没有返回 token');
        return;
      }
      const nextUser = {
        id: data.id,
        username: data.username || values.username || values.email,
        email: data.email || values.email,
      };
      window.localStorage.setItem('token', token);
      window.localStorage.setItem('user', JSON.stringify(nextUser));
      setUser(nextUser);
      setAuthOpen(false);
      form.resetFields();
      message.success(mode === 'login' ? '登录成功' : '注册成功');
    } catch (error: any) {
      message.error(error?.message || '操作失败');
    }
  };

  const logout = () => {
    window.localStorage.removeItem('token');
    window.localStorage.removeItem('user');
    setUser(null);
    router.push('/');
    message.success('已退出登录');
  };

  return (
    <header className="sticky top-0 z-50 border-b border-[rgba(24,32,51,0.1)] bg-[#fffaf1]/82 backdrop-blur-xl">
      <div className="shell flex h-16 items-center justify-between">
        <Link href="/" className="flex items-center gap-3">
          <div className="grid h-10 w-10 place-items-center rounded-lg bg-[#182033] text-lg font-black text-white">
            面
          </div>
          <div className="leading-tight">
            <div className="text-base font-black tracking-tight text-[#182033]">面试舱</div>
            <div className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[#667085]">
              Interview OS
            </div>
          </div>
        </Link>

        <nav className="hidden items-center gap-1 lg:flex">
          {navItems.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`rounded-lg px-3 py-2 text-sm font-semibold transition ${
                  active ? 'bg-white text-[#182033] shadow-sm' : 'text-[#182033]/68 hover:bg-white/70 hover:text-[#182033]'
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="flex items-center gap-2">
          {user ? (
            <Dropdown
              trigger={['click']}
              menu={{
                items: [
                  {
                    key: 'center',
                    icon: <UserOutlined />,
                    label: <Link href="/user/center">个人中心</Link>,
                  },
                  {
                    key: 'models',
                    icon: <SettingOutlined />,
                    label: <Link href="/user/models">模型配置</Link>,
                  },
                  { type: 'divider' },
                  {
                    key: 'logout',
                    danger: true,
                    icon: <LogoutOutlined />,
                    label: <button onClick={logout}>退出登录</button>,
                  },
                ],
              }}
            >
              <Button icon={<UserOutlined />}>{user.username || user.email || '用户'}</Button>
            </Dropdown>
          ) : (
            <Button
              className="ink-button"
              icon={<LoginOutlined />}
              onClick={() => {
                setMode('login');
                setAuthOpen(true);
              }}
            >
              登录
            </Button>
          )}
          <Dropdown
            trigger={['click']}
            menu={{
              items: navItems.map((item) => ({
                key: item.href,
                label: <Link href={item.href}>{item.label}</Link>,
              })),
            }}
          >
            <Button className="lg:hidden" icon={<MenuOutlined />} />
          </Dropdown>
        </div>
      </div>

      <Modal
        title={mode === 'login' ? '登录面试舱' : '创建账号'}
        open={authOpen}
        footer={null}
        onCancel={() => setAuthOpen(false)}
        destroyOnClose
      >
        <Form layout="vertical" form={form} onFinish={submitAuth}>
          {mode === 'register' && (
            <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
              <Input placeholder="输入用户名" />
            </Form.Item>
          )}
          <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
            <Input placeholder="name@example.com" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="输入密码" />
          </Form.Item>
          <Button className="ink-button h-10 w-full" htmlType="submit" icon={<RocketOutlined />}>
            {mode === 'login' ? '登录' : '注册并进入'}
          </Button>
          <button
            type="button"
            className="mt-4 w-full text-sm font-semibold text-[#3157b7]"
            onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
          >
            {mode === 'login' ? '没有账号？去注册' : '已有账号？去登录'}
          </button>
        </Form>
      </Modal>
    </header>
  );
}
