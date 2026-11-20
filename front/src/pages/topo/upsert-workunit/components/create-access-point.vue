<template>
  <div>
    <!-- 接入点名称 -->
    <Form.FormItem
      :label="$t('topoManager.workUnit.form.accessPointName')"
      :property="`accesspoint_name.${index}`"
      :rules="accessPointNameRule"
      label-width="120">
      <Input v-model="data.accesspoint_name" class="w-[283px]" />
    </Form.FormItem>
    <!-- cluster -->
    <Form.FormItem
      label="cluster"
      :property="`cluster.${index}`"
      :rules="clusterRule"
      label-width="120">
      <InputGroup v-model:values="data.endpoints.cluster" :placeholder="inputPlaceholder"></InputGroup>
    </Form.FormItem>
    <!-- file -->
    <Form.FormItem
      label="file"
      :property="`file.${index}`"
      :rules="fileRule"
      label-width="120">
      <InputGroup v-model:values="data.endpoints.file" :placeholder="$t('topoManager.workUnit.form.input.downstreamFilePlaceholder')"></InputGroup>
    </Form.FormItem>
    <!-- data -->
    <Form.FormItem
      label="data"
      :property="`data.${index}`"
      :rules="dataRule"
      label-width="120"
      :class="cloudSgEnabled ? '' : 'mb-0'">
      <InputGroup v-model:values="data.endpoints.data" :placeholder="inputPlaceholder"></InputGroup>
    </Form.FormItem>
    <template>
      <!-- 云网络策略 -->
      <Form.FormItem
        :label="$t('topoManager.workUnit.form.cloudSecurityGroup')"
        label-width="120"
        :class="cloudSgEnabled || data.cloud_security_group_pool_id ? '' : 'mb-0'">
        <Switcher v-model="cloudSgEnabled" :disabled="!hasCloudSgManagePermission" size="small" />
        <span class="ml-[8px] text-[#979BA5] text-[12px]">
          {{ $t('topoManager.workUnit.form.cloudSecurityGroupTips') }}
        </span>
      </Form.FormItem>
      <Form.FormItem
        v-if="!hasCloudSgManagePermission && data.cloud_security_group_pool_id"
        :label="$t('topoManager.workUnit.form.cloudSgPool')"
        label-width="120"
        class="mb-0">
        <Input :model-value="data.cloud_security_group_pool_id" class="w-[283px]" disabled />
      </Form.FormItem>
      <Form.FormItem
        v-else-if="hasCloudSgManagePermission && cloudSgEnabled"
        :label="$t('topoManager.workUnit.form.cloudSgPool')"
        :property="`cloudSgPool.${index}`"
        :rules="cloudSgPoolRule"
        label-width="120"
        class="mb-0">
        <Select
          v-model="data.cloud_security_group_pool_id"
          class="w-[283px]"
          :placeholder="$t('topoManager.workUnit.form.cloudSgPoolPlaceholder')">
          <Select.Option
            v-for="pool in poolOptions"
            :id="pool.pool_id"
            :key="pool.pool_id"
            :name="pool.pool_name">
          </Select.Option>
        </Select>
      </Form.FormItem>
    </template>
  </div>
</template>

<script lang="ts" setup>
import { Form, Input, Select, Switcher } from 'bkui-vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import InputGroup from './input-group.vue';

import type { CloudSecurityGroupPool } from '@/@types/topo';

const data = defineModel<Omit<AccessPoint, 'accesspoint_id' | 'tenant_id'>>('data', { required: true });

defineProps({
  index: {
    type: Number,
    required: true,
  },
  hasCloudSgManagePermission: {
    type: Boolean,
    default: false,
  },
  poolOptions: {
    type: Array as () => CloudSecurityGroupPool[],
    default: () => [],
  },
});

const enabledLocally = ref(Boolean(data.value.cloud_security_group_pool_id));
const cloudSgEnabled = computed<boolean>({
  get: () => Boolean(data.value.cloud_security_group_pool_id) || enabledLocally.value,
  set: (enabled: boolean) => {
    enabledLocally.value = enabled;
    if (!enabled) data.value.cloud_security_group_pool_id = '';
  },
});
watch(() => data.value, (value) => {
  enabledLocally.value = Boolean(value.cloud_security_group_pool_id);
});

const { t } = useI18n();
const inputPlaceholder = t('topoManager.workUnit.form.input.placeholder');

// 接入点名称校验规则
const accessPointNameRule = [{
  trigger: 'blur',
  validator: () => data.value.accesspoint_name !== '',
}];

// cluster校验规则
const clusterRule = [{
  trigger: 'blur',
  validator: () => data.value.endpoints.cluster.every(item => item !== ''),
},
{
  trigger: 'blur',
  validator: () => data.value.endpoints.cluster.every(item => isInputValid(item)),
  message: t('topoManager.workUnit.form.input.validate'),
}];

// file校验规则
const fileRule = [{
  trigger: 'blur',
  validator: () => data.value.endpoints.file.every(item => item !== ''),
},
{
  trigger: 'blur',
  validator: () => data.value.endpoints.file.every(item => isInputValid(item)),
  message: t('topoManager.workUnit.form.input.validate'),
}];

// data校验规则
const dataRule = [{
  trigger: 'blur',
  validator: () => data.value.endpoints.data.every(item => item !== ''),
},
{
  trigger: 'blur',
  validator: () => data.value.endpoints.data.every(item => isInputValid(item)),
  message: t('topoManager.workUnit.form.input.validate'),
}];

// 开启云网络策略后必须选择一个全局安全组池
const cloudSgPoolRule = [{
  trigger: 'change',
  validator: () => !cloudSgEnabled.value || data.value.cloud_security_group_pool_id !== '',
}];

const isInputValid = (value: string) => {
  if (!value) return false;
  if (!/^\d{1,3}(\.\d{1,3}){3}:\d+$/.test(value)) {
    return false;
  }

  const validateIP = (ip: string) => ip.split('.').every((segment) => {
    const num = parseInt(segment, 10);
    return num >= 0 && num <= 255 && segment === num.toString();
  });

  const validatePort = (port: string) => {
    const portNum = parseInt(port, 10);
    return portNum >= 0 && portNum <= 65535;
  };

  const [ip, port] = value.split(':');
  return validateIP(ip) && validatePort(port);
};

</script>
