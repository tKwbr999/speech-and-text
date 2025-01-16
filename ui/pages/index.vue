<template>
  <div>
    <VoiceInput />
  </div>
</template>

<script lang="ts">
import Vue from 'vue'
import VoiceInput from '../components/VoiceInput.vue';

export default Vue.extend({
  name: 'IndexPage',
  components: {
    VoiceInput,
  },
})
</script>
