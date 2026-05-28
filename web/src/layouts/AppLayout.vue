<script setup lang="ts">
import { Activity, LogOut, Package, UserCircle, Users } from '@lucide/vue';
import {
  NButton,
  NIcon,
  NLayout,
  NLayoutContent,
  NLayoutSider,
  NMenu,
  NSpace,
  type MenuOption
} from 'naive-ui';
import { computed, h, onMounted } from 'vue';
import { RouterView, useRoute, useRouter } from 'vue-router';

import { useSessionStore } from '../stores/session';

const route = useRoute();
const router = useRouter();
const session = useSessionStore();

function icon(component: typeof Users) {
  return () => h(NIcon, null, { default: () => h(component) });
}

const menuOptions: MenuOption[] = [
  {
    label: '团队',
    key: 'teams',
    icon: icon(Users)
  }
];

const selectedMenu = computed(() => {
  if (route.path.includes('/products/')) return 'teams';
  return 'teams';
});

onMounted(async () => {
  if (!session.user && session.isAuthenticated) {
    await session.loadCurrentUser().catch(() => session.clearSession());
  }
});

function handleMenu(key: string) {
  if (key === 'teams') {
    router.push('/app/teams');
  }
}

function logout() {
  session.clearSession();
  router.push('/login');
}
</script>

<template>
  <n-layout has-sider class="app-shell">
    <n-layout-sider
      class="sidebar"
      :width="232"
      :collapsed-width="72"
      collapse-mode="width"
      show-trigger="bar"
      bordered
    >
      <div class="sidebar-header">
        <div class="brand-mark">
          <Activity :size="20" />
        </div>
        <span class="sidebar-title">PulseRoad</span>
      </div>
      <n-menu
        :value="selectedMenu"
        :options="menuOptions"
        :collapsed-icon-size="20"
        @update:value="handleMenu"
      />
    </n-layout-sider>

    <n-layout>
      <header class="topbar">
        <div class="topbar-title">
          <h1>产品团队工作台</h1>
          <p>管理团队和团队下的产品</p>
        </div>
        <n-space align="center" :wrap="false">
          <n-icon :size="20">
            <UserCircle />
          </n-icon>
          <span>{{ session.user?.name || '当前用户' }}</span>
          <n-button quaternary size="small" @click="logout">
            <template #icon>
              <n-icon><LogOut /></n-icon>
            </template>
            退出
          </n-button>
        </n-space>
      </header>

      <n-layout-content>
        <RouterView />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>
