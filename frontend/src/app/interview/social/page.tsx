'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Alert, Button, Form, Input, Select, Skeleton, Tag, message } from 'antd';
import {
  BankOutlined,
  CodeOutlined,
  CompassOutlined,
  FileTextOutlined,
  PlayCircleOutlined,
  ProfileOutlined,
} from '@ant-design/icons';
import { ResumeItem, formatFileSize, resumeService } from '@/services/api/resume';

type InterviewFormValues = {
  resume_id: number;
  position: string;
  company?: string;
  difficulty: string;
};

const difficultyOptions = [
  { value: '简单', label: '简单' },
  { value: '中等', label: '中等' },
  { value: '困难', label: '困难' },
];

export default function SocialInterviewPage() {
  const [form] = Form.useForm<InterviewFormValues>();
  const [resumes, setResumes] = useState<ResumeItem[]>([]);
  const [loadingResumes, setLoadingResumes] = useState(true);
  const [starting, setStarting] = useState(false);
  const router = useRouter();

  const selectedResumeId = Form.useWatch('resume_id', form);
  const selectedResume = useMemo(
    () => resumes.find((item) => item.id === selectedResumeId),
    [resumes, selectedResumeId],
  );

  const fetchResumes = async () => {
    setLoadingResumes(true);
    try {
      const data = await resumeService.list();
      setResumes(data.resumes);
      const defaultResume = data.resumes.find((item) => item.is_default === 1) || data.resumes[0];
      if (defaultResume) form.setFieldsValue({ resume_id: defaultResume.id });
    } catch (error: any) {
      message.error(error?.message || '简历列表加载失败');
    } finally {
      setLoadingResumes(false);
    }
  };

  useEffect(() => {
    void fetchResumes();
  }, []);

  const handleStart = async () => {
    try {
      const values = await form.validateFields();
      const params = {
        domain: '社招简历面试',
        difficulty: values.difficulty,
        position: values.position,
        company: values.company || '',
        resume_id: values.resume_id,
      };
      (window as any).__interviewParams = params;
      window.sessionStorage.setItem('interviewParams', JSON.stringify(params));
      setStarting(true);
      router.push('/interview/social/start');
    } catch {
      message.warning('请先补全面试配置');
    }
  };

  return (
    <div className="shell py-8">
      <section className="grid gap-5 lg:grid-cols-[1fr_440px]">
        <div className="panel rounded-lg p-8">
          <div className="mb-10 max-w-3xl">
            <div className="mb-3 inline-flex items-center gap-2 rounded-lg bg-[#182033] px-3 py-2 text-sm font-semibold text-white">
              <CompassOutlined />
              社招综合面试
            </div>
            <h1 className="m-0 text-4xl font-black leading-tight text-[#182033]">
              用项目经历和岗位目标，模拟一场更接近真实社招的追问。
            </h1>
            <p className="mt-4 text-base leading-7 text-[#667085]">
              直接从后端简历列表中选择上下文。启动后前端调用 POST /mianshi/stream/start，
              后端通过 SSE 连续发送问题。
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-3">
            {[
              { icon: <ProfileOutlined />, title: '简历驱动', desc: '围绕项目、经历和岗位匹配点追问。' },
              { icon: <CodeOutlined />, title: '技术深挖', desc: '覆盖方案取舍、性能、稳定性和工程落地。' },
              { icon: <BankOutlined />, title: '岗位贴合', desc: '输入目标公司和岗位，让问题更聚焦。' },
            ].map((item) => (
              <div key={item.title} className="surface rounded-lg p-5">
                <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg bg-[#d96c4a] text-lg text-white">
                  {item.icon}
                </div>
                <div className="font-bold text-[#182033]">{item.title}</div>
                <p className="m-0 mt-2 text-sm leading-6 text-[#667085]">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="panel rounded-lg bg-white p-6">
          <h2 className="m-0 mb-1 text-xl font-black text-[#182033]">面试配置</h2>
          <p className="mb-6 text-sm text-[#667085]">选择简历后进入实时答题页面。</p>

          {loadingResumes ? (
            <Skeleton active paragraph={{ rows: 6 }} />
          ) : resumes.length === 0 ? (
            <Alert
              type="warning"
              showIcon
              message="还没有简历"
              description={
                <span>
                  请先去 <Link href="/resume" className="font-bold text-[#3157b7]">简历库</Link> 上传并解析简历。
                </span>
              }
            />
          ) : (
            <Form form={form} layout="vertical" initialValues={{ position: '后端开发工程师', difficulty: '中等' }}>
              <Form.Item label="选择简历" name="resume_id" rules={[{ required: true, message: '请选择简历' }]}>
                <Select
                  optionLabelProp="label"
                  options={resumes.map((resume) => ({
                    value: resume.id,
                    label: resume.file_name,
                  }))}
                />
              </Form.Item>

              {selectedResume && (
                <div className="mb-5 rounded-lg border border-[rgba(24,32,51,0.1)] bg-[#fffaf1] p-4">
                  <div className="mb-2 flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2 font-bold text-[#182033]">
                      <FileTextOutlined />
                      {selectedResume.file_name}
                    </div>
                    {selectedResume.is_default === 1 && <Tag color="green">默认</Tag>}
                  </div>
                  <div className="text-xs text-[#667085]">
                    ID: {selectedResume.id} · {selectedResume.file_type || 'file'} · {formatFileSize(selectedResume.file_size)}
                  </div>
                </div>
              )}

              <Form.Item label="目标岗位" name="position" rules={[{ required: true, message: '请输入目标岗位' }]}>
                <Input placeholder="后端开发工程师" />
              </Form.Item>
              <Form.Item label="目标公司" name="company">
                <Input placeholder="可选，例如 字节跳动" />
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
                开始社招面试
              </Button>
            </Form>
          )}
        </div>
      </section>
    </div>
  );
}
