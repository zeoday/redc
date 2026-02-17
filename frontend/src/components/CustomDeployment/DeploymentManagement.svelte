<script lang="ts">
  import { onMount } from 'svelte';
  import CustomDeploymentList from './CustomDeploymentList.svelte';

  let { t } = $props();

  let listComponent: any;

  function handleSelectDeployment(deployment: any) {
    // 不需要处理，展开逻辑在 CustomDeploymentList 中
  }

  function handleRefresh() {
    if (listComponent) {
      listComponent.refresh();
    }
  }

  // Periodically check batch selection state
  onMount(() => {
    // 可以在这里添加定期刷新逻辑
  });
</script>

<div class="deployment-management">
  <div class="management-content">
    <CustomDeploymentList 
      bind:this={listComponent}
      t={t}
      onSelectDeployment={handleSelectDeployment}
      onRefresh={handleRefresh}
    />
  </div>
</div>

<style>
  .deployment-management {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .management-content {
    flex: 1;
    min-height: 0;
  }
</style>
