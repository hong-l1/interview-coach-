'use client';

import { useEffect, useMemo, useState } from 'react';
import { Button, Form, Input, Modal, Popconfirm, Select, Switch, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  ApiOutlined,
  CloudServerOutlined,
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import apiClient from '@/services/api/client';

type ModelItem = {
  id: number;
  name: string;
  model_name: string;
  base_url: string;
  provider_name: string;
  is_default: number;
  protocol?: string;
  created_at?: number | string;
};

type ModelFormValues = {
  name: string;
  model_name: string;
  base_url: string;
  api_key: string;
  provider_name: string;
  is_default: boolean;
};

function normalizeModel(raw: any): ModelItem {
  return {
    id: Number(raw?.id ?? raw?.ID ?? 0),
    name: raw?.name ?? raw?.Name ?? '',
    model_name: raw?.model_name ?? raw?.modelKey ?? raw?.model_key ?? raw?.ModelKey ?? '',
    base_url: raw?.base_url ?? raw?.baseURL ?? raw?.BaseURL ?? '',
    provider_name: raw?.provider_name ?? raw?.providerName ?? raw?.ProviderName ?? '',
    is_default: Number(raw?.is_default ?? raw?.IsDefault ?? 0),
    protocol: raw?.protocol ?? raw?.Protocol ?? 'openai',
    created_at: raw?.created_at ?? raw?.createdAt ?? raw?.CreatedAt,
  };
}

function buildPayload(values: ModelFormValues) {
  return {
    name: values.name,
    model_name: values.model_name,
    base_url: values.base_url,
    api_key: values.api_key,
    provider_name: values.provider_name,
    is_default: values.is_default ? 1 : 0,
  };
}

export default function UserModelsPage() {
  const [list, setList] = useState<ModelItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<ModelItem | null>(null);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm<ModelFormValues>();

  const fetchList = async () => {
    setLoading(true);
    try {
      const res: any = await apiClient.get('/user/models');
      const rows = Array.isArray(res) ? res : Array.isArray(res?.list) ? res.list : [];
      setList(rows.map(normalizeModel));
    } catch (error: any) {
      message.error(error?.message || '模型列表加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchList();
  }, []);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({
      provider_name: 'OpenAI',
      is_default: list.length === 0,
    } as Partial<ModelFormValues>);
    setModalOpen(true);
  };

  const openEdit = async (row: ModelItem) => {
    try {
      const detail = normalizeModel(await apiClient.get(`/user/model/${row.id}`));
      setEditing(detail);
      form.setFieldsValue({
        name: detail.name,
        model_name: detail.model_name,
        base_url: detail.base_url,
        provider_name: detail.provider_name || 'OpenAI',
        api_key: '',
        is_default: detail.is_default === 1,
      });
      setModalOpen(true);
    } catch (error: any) {
      message.error(error?.message || '模型详情加载失败');
    }
  };

  const saveModel = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);
      const payload = buildPayload(values);
      if (editing) {
        await apiClient.put(`/user/model/${editing.id}`, payload);
        message.success('模型已更新');
      } else {
        await apiClient.post('/user/model', payload);
        message.success('模型已创建');
      }
      setModalOpen(false);
      await fetchList();
    } catch (error: any) {
      if (error?.errorFields) return;
      message.error(error?.message || '保存失败');
    } finally {
      setSaving(false);
    }
  };

  const deleteModel = async (id: number) => {
    try {
      await apiClient.delete(`/user/model/${id}`);
      message.success('模型已删除');
      await fetchList();
    } catch (error: any) {
      message.error(error?.message || '删除失败');
    }
  };

  const columns = useMemo<ColumnsType<ModelItem>>(
    () => [
      {
        title: '名称',
        dataIndex: 'name',
        render: (value: string, row) => (
          <div>
            <div className="font-bold text-[#172033]">{value || '-'}</div>
            <div className="text-xs text-[#5f6878]">ID: {row.id}</div>
          </div>
        ),
      },
      {
        title: '模型',
        dataIndex: 'model_name',
        render: (value: string) => <Tag className="font-mono">{value || '-'}</Tag>,
      },
      {
        title: '供应商',
        dataIndex: 'provider_name',
        render: (value: string) => value || 'OpenAI',
      },
      {
        title: 'Base URL',
        dataIndex: 'base_url',
        ellipsis: true,
      },
      {
        title: '默认',
        dataIndex: 'is_default',
        width: 100,
        render: (value: number) => (value === 1 ? <Tag color="green">默认</Tag> : <Tag>备用</Tag>),
      },
      {
        title: '操作',
        width: 130,
        render: (_, row) => (
          <div className="flex gap-2">
            <Button icon={<EditOutlined />} onClick={() => openEdit(row)} />
            <Popconfirm title="确认删除这个模型？" onConfirm={() => deleteModel(row.id)}>
              <Button danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </div>
        ),
      },
    ],
    [],
  );

  return (
    <div className="shell py-8">
      <section className="panel rounded-lg p-6">
        <div className="mb-6 grid gap-5 lg:grid-cols-[1fr_300px]">
          <div>
            <div className="mb-2 inline-flex items-center gap-2 rounded-lg bg-[#172033] px-3 py-2 text-sm font-semibold text-white">
              <CloudServerOutlined />
              Model Provider
            </div>
            <h1 className="m-0 text-3xl font-black text-[#172033]">模型配置</h1>
            <p className="m-0 mt-2 text-sm text-[#5f6878]">
              已对齐后端路由：GET /user/models、POST /user/model、PUT/DELETE /user/model/:id。
            </p>
          </div>
          <div className="surface rounded-lg p-4">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <div className="text-2xl font-black text-[#172033]">{list.length}</div>
                <div className="text-xs text-[#5f6878]">模型数量</div>
              </div>
              <div>
                <div className="text-2xl font-black text-[#172033]">{list.some((item) => item.is_default === 1) ? '已设' : '未设'}</div>
                <div className="text-xs text-[#5f6878]">默认模型</div>
              </div>
            </div>
            <div className="mt-4 flex gap-2">
              <Button icon={<ReloadOutlined />} onClick={fetchList} className="flex-1">
                刷新
              </Button>
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreate} className="ink-button flex-1">
                新增
              </Button>
            </div>
          </div>
        </div>

        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={list}
          pagination={false}
          scroll={{ x: 820 }}
        />
      </section>

      <Modal
        title={
          <div className="flex items-center gap-2">
            <ApiOutlined />
            {editing ? '编辑模型' : '新增模型'}
          </div>
        }
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={saveModel}
        okText={editing ? '更新' : '创建'}
        confirmLoading={saving}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            provider_name: 'OpenAI',
            is_default: false,
          }}
        >
          <Form.Item label="显示名称" name="name" rules={[{ required: true, message: '请输入显示名称' }]}>
            <Input placeholder="例如 我的 GPT-4o" />
          </Form.Item>
          <Form.Item label="模型名" name="model_name" rules={[{ required: true, message: '请输入模型名' }]}>
            <Input placeholder="例如 gpt-4o-mini" />
          </Form.Item>
          <Form.Item label="Base URL" name="base_url" rules={[{ required: true, message: '请输入 Base URL' }]}>
            <Input placeholder="例如 https://api.openai.com/v1" />
          </Form.Item>
          <Form.Item label="API Key" name="api_key" rules={[{ required: true, message: '请输入 API Key' }]}>
            <Input.Password placeholder="后端更新接口当前要求每次都传 API Key" autoComplete="new-password" />
          </Form.Item>
          <Form.Item label="供应商" name="provider_name" rules={[{ required: true, message: '请选择供应商' }]}>
            <Select
              options={[
                { value: 'OpenAI', label: 'OpenAI' },
                { value: 'DeepSeek', label: 'DeepSeek' },
                { value: '豆包', label: '豆包' },
                { value: '本地模型', label: '本地模型' },
              ]}
            />
          </Form.Item>
          <Form.Item label="设为默认" name="is_default" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
