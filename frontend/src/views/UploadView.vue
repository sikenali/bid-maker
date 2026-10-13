<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import TopNav from '@/components/layout/TopNav.vue';
import PageFooter from '@/components/layout/PageFooter.vue';
import UploadCard from '@/components/upload/UploadCard.vue';
import CreditQualificationModal from '@/components/layout/CreditQualificationModal.vue';
import BiddingSitesModal from '@/components/layout/BiddingSitesModal.vue';

const router = useRouter();
const showCreditModal = ref(false);
const showBiddingModal = ref(false);

function handleSuccess(jobId: string) {
  try { sessionStorage.setItem('bidextractor.lastJobId', jobId); } catch {}
  router.replace({ path: '/result' });
}

function handleReplace() {
  try { sessionStorage.removeItem('bidextractor.lastJobId'); } catch {}
}
</script>

<template>
  <div class="upload-view">
    <TopNav title="智能标书解决方案平台-提取" :show-back="false" />

    <div class="upload-body">
      <UploadCard @success="handleSuccess" @replace="handleReplace" />
    </div>

    <PageFooter @credit-query="showCreditModal = true" @bidding-sites="showBiddingModal = true" />

    <CreditQualificationModal v-model:visible="showCreditModal" />
    <BiddingSitesModal v-model:visible="showBiddingModal" />
  </div>
</template>

<style scoped>
.upload-view {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--color-bg-main, #FAF7F2);
}
.upload-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0 32px;
}
</style>
