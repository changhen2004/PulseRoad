<script setup lang="ts">
import { Plus } from '@lucide/vue';
import {
  NButton,
  NDataTable,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NSpace,
  NTag,
  type DataTableColumns,
  useMessage
} from 'naive-ui';
import { computed, h, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { teamsApi } from '../api/teams';
import type { Team } from '../api/types';

const router = useRouter();
const message = useMessage();
const loading = ref(false);
const saving = ref(false);
const drawerOpen = ref(false);
const teams = ref<Team[]>([]);

const form = reactive({
  name: '',
  description: ''
});

const canCreate = computed(() => form.name.trim().length > 0);

const columns: DataTableColumns<Team> = [
  {
    title: '团队',
    key: 'name',
    render(row) {
      return h('div', { class: 'table-title-cell' }, [
        h('strong', row.name),
        row.description ? h('span', row.description) : null
      ]);
    }
  },
  {
    title: '角色',
    key: 'role',
    width: 120,
    render(row) {
      return h(NTag, { type: row.role === 'owner' ? 'success' : 'info', bordered: false }, () =>
        row.role
      );
    }
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render(row) {
      return formatDate(row.created_at);
    }
  },
  {
    title: '',
    key: 'actions',
    width: 110,
    align: 'right',
    render(row) {
      return h(
        NButton,
        {
          size: 'small',
          onClick: () => router.push(`/app/teams/${row.id}`)
        },
        () => '查看'
      );
    }
  }
];

onMounted(loadTeams);

async function loadTeams() {
  loading.value = true;
  try {
    teams.value = await teamsApi.list();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载团队失败');
  } finally {
    loading.value = false;
  }
}

async function createTeam() {
  if (!canCreate.value || saving.value) return;
  saving.value = true;
  try {
    const team = await teamsApi.create({
      name: form.name.trim(),
      description: form.description.trim()
    });
    drawerOpen.value = false;
    form.name = '';
    form.description = '';
    await loadTeams();
    router.push(`/app/teams/${team.id}`);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '创建团队失败');
  } finally {
    saving.value = false;
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).format(new Date(value));
}

function teamRowKey(row: Team) {
  return row.id;
}
</script>

<template>
  <main class="page">
    <div class="page-header">
      <div>
        <h2 class="page-title">我的团队</h2>
        <p class="page-description">查看你已加入的团队，创建团队后你会自动成为 owner。</p>
      </div>
      <n-button type="primary" @click="drawerOpen = true">
        <template #icon>
          <n-icon><Plus /></n-icon>
        </template>
        新建团队
      </n-button>
    </div>

    <section class="content-panel">
      <n-data-table
        :columns="columns"
        :data="teams"
        :loading="loading"
        :bordered="false"
        :single-line="false"
        :row-key="teamRowKey"
      />
      <div v-if="!loading && teams.length === 0" class="empty-state">还没有团队，先创建一个。</div>
    </section>

    <n-drawer v-model:show="drawerOpen" :width="420" placement="right">
      <n-drawer-content title="新建团队" closable>
        <n-form label-placement="top" @submit.prevent="createTeam">
          <n-form-item label="团队名称">
            <n-input v-model:value="form.name" placeholder="例如：核心产品组" />
          </n-form-item>
          <n-form-item label="描述">
            <n-input
              v-model:value="form.description"
              type="textarea"
              placeholder="团队负责的产品方向"
              :autosize="{ minRows: 4, maxRows: 8 }"
            />
          </n-form-item>
          <n-space justify="end">
            <n-button @click="drawerOpen = false">取消</n-button>
            <n-button type="primary" attr-type="submit" :disabled="!canCreate" :loading="saving">
              创建
            </n-button>
          </n-space>
        </n-form>
      </n-drawer-content>
    </n-drawer>
  </main>
</template>
