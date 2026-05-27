'use client';

import Link from 'next/link';
import { Alert, Button, Empty, Popconfirm, Skeleton, Tag, Upload, message } from 'antd';
import type { UploadProps } from 'antd';
import {
  DeleteOutlined,
  FileSearchOutlined,
  FileTextOutlined,
  InboxOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  StarFilled,
  StarOutlined,
} from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { ResumeItem, formatFileSize, formatTimestamp, resumeService } from '@/services/api/resume';

export default function ResumePage() {
  const [uploading, setUploading] = useState(false);
  const [loading, setLoading] = useState(true);
  const [resumes, setResumes] = useState<ResumeItem[]>([]);

  const fetchResumes = async () => {
    setLoading(true);
    try {
      const data = await resumeService.list();
      setResumes(data.resumes);
    } catch (error: any) {
      message.error(error?.message || '简历列表加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchResumes();
  }, []);

  const beforeUpload: UploadProps['beforeUpload'] = async (file) => {
    setUploading(true);
    try {
      await resumeService.upload(file);
      message.success('简历上传解析成功');
      await fetchResumes();
    } catch (error: any) {
      message.error(error?.message || '上传失败');
    } finally {
      setUploading(false);
    }
    return false;
  };

  const setDefault = async (id: number) => {
    try {
      await resumeService.setDefault(id);
      message.success('默认简历已更新');
      await fetchResumes();
    } catch (error: any) {
      message.error(error?.message || '设置失败');
    }
  };

  const removeResume = async (id: number) => {
    try {
      await resumeService.remove(id);
      message.success('简历已删除');
      await fetchResumes();
    } catch (error: any) {
      message.error(error?.message || '删除失败');
    }
  };

  return (
    <div className="shell py-8">
      <section className="grid gap-5 lg:grid-cols-[420px_1fr]">
        <aside className="panel h-fit rounded-lg p-6">
          <div className="mb-6">
            <div className="mb-3 inline-flex items-center gap-2 rounded-lg bg-[#3157b7] px-3 py-2 text-sm font-semibold text-white">
              <FileSearchOutlined />
              Resume Library
            </div>
            <h1 className="m-0 text-3xl font-black text-[#182033]">简历库</h1>
            <p className="mt-3 text-sm leading-6 text-[#667085]">
              上传后端会解析并入库，社招和校招入口会直接从这里选择简历作为面试上下文。
            </p>
          </div>

          <Upload.Dragger
            name="file"
            accept=".pdf,.doc,.docx,.txt"
            showUploadList={false}
            beforeUpload={beforeUpload}
            disabled={uploading}
            className="!rounded-lg"
          >
            <p className="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p className="ant-upload-text">{uploading ? '解析中...' : '点击或拖拽文件上传'}</p>
            <p className="ant-upload-hint">支持 PDF、DOC、DOCX、TXT</p>
          </Upload.Dragger>

          <Alert
            className="mt-5"
            type="info"
            showIcon
            message="建议设置默认简历"
            description="社招和校招入口会优先选中默认简历，没有默认时选中最新上传的一份。"
          />
        </aside>

        <main className="panel rounded-lg p-6">
          <div className="mb-6 flex flex-col justify-between gap-4 md:flex-row md:items-center">
            <div>
              <h2 className="m-0 text-2xl font-black text-[#182033]">已上传简历</h2>
              <p className="m-0 mt-1 text-sm text-[#667085]">共 {resumes.length} 份</p>
            </div>
            <Button icon={<ReloadOutlined />} onClick={fetchResumes}>
              刷新
            </Button>
          </div>

          {loading ? (
            <Skeleton active paragraph={{ rows: 8 }} />
          ) : resumes.length === 0 ? (
            <Empty description="还没有上传简历" image={Empty.PRESENTED_IMAGE_SIMPLE} />
          ) : (
            <div className="grid gap-4">
              {resumes.map((resume) => (
                <div key={resume.id} className="surface rounded-lg p-5 transition hover:-translate-y-0.5 hover:bg-white">
                  <div className="flex flex-col justify-between gap-4 md:flex-row md:items-start">
                    <div className="flex gap-4">
                      <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-[#fff4e8] text-xl text-[#d96c4a]">
                        <FileTextOutlined />
                      </div>
                      <div>
                        <div className="flex flex-wrap items-center gap-2">
                          <h3 className="m-0 text-lg font-black text-[#182033]">{resume.file_name}</h3>
                          {resume.is_default === 1 && <Tag color="green">默认</Tag>}
                        </div>
                        <div className="mt-2 text-sm text-[#667085]">
                          ID: {resume.id} · {resume.file_type || 'file'} · {formatFileSize(resume.file_size)} · 上传于 {formatTimestamp(resume.created_at)}
                        </div>
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      <Button
                        icon={resume.is_default === 1 ? <StarFilled /> : <StarOutlined />}
                        disabled={resume.is_default === 1}
                        onClick={() => setDefault(resume.id)}
                      >
                        {resume.is_default === 1 ? '已默认' : '设默认'}
                      </Button>
                      <Popconfirm title="确认删除这份简历？" onConfirm={() => removeResume(resume.id)}>
                        <Button danger icon={<DeleteOutlined />}>
                          删除
                        </Button>
                      </Popconfirm>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="mt-6 flex flex-wrap gap-3">
            <Link href="/interview/campus">
              <Button type="primary" icon={<PlayCircleOutlined />} className="ink-button">
                去校招面试
              </Button>
            </Link>
            <Link href="/interview/social">
              <Button>去社招面试</Button>
            </Link>
          </div>
        </main>
      </section>
    </div>
  );
}
