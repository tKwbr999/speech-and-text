<template>
  <div class="audio-recorder">
    <button 
      @click="startRecording" 
      :disabled="isRecording"
      class="record-button"
    >
      {{ isRecording ? '録音中...' : '録音開始' }}
    </button>
    <button 
      @click="stopRecording" 
      :disabled="!isRecording"
      class="stop-button"
    >
      録音停止
    </button>
    
    <div v-if="recordingDuration" class="duration">
      録音時間: {{ formatDuration(recordingDuration) }}
    </div>

    <div v-if="audioBlob" class="playback-control">
      <audio ref="audioPlayer" controls></audio>
    </div>

    <p v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </p>

    <div v-if="audioBlob" class="download-buttons">
      <button @click="saveRecording">
        録音データを保存
      </button>
      <button @click="sendToText">
        テキストにする
      </button>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      isRecording: false,
      mediaRecorder: null,
      audioChunks: [],
      audioBlob: null,
      errorMessage: '',
      stream: null,
      recordingStartTime: null,
      recordingDuration: 0,
      durationTimer: null,
      audioContext: null,
      mediaStreamDestination: null,
      mediaStreamSource: null
    };
  },
  methods: {
    async initializeRecorder() {
      try {
        // オーディオコンテキストの初期化
        this.audioContext = new (window.AudioContext || window.webkitAudioContext)({
          sampleRate: 44100,
        });

        // マイクからの入力を取得
        this.stream = await navigator.mediaDevices.getUserMedia({ 
          audio: {
            channelCount: 1,
            echoCancellation: true,
            noiseSuppression: true,
            autoGainControl: true,
            sampleRate: 44100,
          }
        });

        // MediaStreamSourceNodeの作成
        this.mediaStreamSource = this.audioContext.createMediaStreamSource(this.stream);
        
        // MediaStreamAudioDestinationNodeの作成
        this.mediaStreamDestination = this.audioContext.createMediaStreamDestination();
        
        // ソースとデスティネーションを接続
        this.mediaStreamSource.connect(this.mediaStreamDestination);

        // MediaRecorderの設定
        this.mediaRecorder = new MediaRecorder(this.mediaStreamDestination.stream, {
          mimeType: 'audio/webm;codecs=opus',
          audioBitsPerSecond: 128000
        });

        this.mediaRecorder.ondataavailable = (event) => {
          if (event.data.size > 0) {
            this.audioChunks.push(event.data);
          }
        };

        this.mediaRecorder.onstop = async () => {
          this.audioBlob = new Blob(this.audioChunks);

          this.$nextTick(() => {
            if (this.$refs.audioPlayer) {
              const audioUrl = URL.createObjectURL(this.audioBlob);
              this.$refs.audioPlayer.src = audioUrl;
              this.$refs.audioPlayer.load();
            }
          });

          this.stopDurationTimer();
          
          // ストリームとオーディオコンテキストのクリーンアップ
          if (this.stream) {
            this.stream.getTracks().forEach(track => track.stop());
          }
          if (this.audioContext && this.audioContext.state !== 'closed') {
            await this.audioContext.close();
          }
        };

        return this.mediaRecorder;

      } catch (error) {
        console.error('Error initializing recorder:', error);
        this.errorMessage = 'マイクへのアクセスに失敗しました。';
        return null;
      }
    },

    async startRecording() {
      try {
        this.mediaRecorder = await this.initializeRecorder();
        
        if (this.mediaRecorder && this.mediaRecorder.state !== 'recording') {
          this.isRecording = true;
          this.audioChunks = [];
          this.audioBlob = null;
          this.errorMessage = '';
          
          this.mediaRecorder.start(1000); // 1秒ごとにデータを取得
          
          this.recordingStartTime = Date.now();
          this.startDurationTimer();
        } else if (!this.mediaRecorder) {
          this.errorMessage = 'レコーダーの初期化に失敗しました。';
        }
      } catch (error) {
        console.error('Error starting recording:', error);
        this.errorMessage = '録音開始に失敗しました。';
        this.isRecording = false;
      }
    },

    stopRecording() {
      if (this.mediaRecorder && this.mediaRecorder.state === 'recording') {
        this.isRecording = false;
        this.mediaRecorder.stop();
      }
    },

    startDurationTimer() {
      this.durationTimer = setInterval(() => {
        this.recordingDuration = Math.floor((Date.now() - this.recordingStartTime) / 1000);
      }, 1000);
    },

    stopDurationTimer() {
      if (this.durationTimer) {
        clearInterval(this.durationTimer);
        this.durationTimer = null;
      }
    },

    formatDuration(seconds) {
      const minutes = Math.floor(seconds / 60);
      const remainingSeconds = seconds % 60;
      return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
    },

    async sendToText() {
      if (!this.audioBlob) return;

      try {
        const formData = new FormData();
        formData.append('audio', this.audioBlob, 'recording.webm');

        const response = await this.$axios.post('http://localhost:8080/api/speech-to-text', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        });

        console.log('Text from audio:', response.data);
      } catch (error) {
        console.error('Error sending audio to text:', error);
        this.errorMessage = 'テキスト変換に失敗しました。';
      }
    },
    saveRecording() {
      if (!this.audioBlob) return;

      // 拡張子は.webmで保存（オーディオ品質を保持するため）
      const filename = `recording_${new Date().toISOString()}.webm`;
      const url = URL.createObjectURL(this.audioBlob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }
  },

  beforeDestroy() {
    this.stopDurationTimer();
    if (this.stream) {
      this.stream.getTracks().forEach(track => track.stop());
    }
    if (this.audioContext && this.audioContext.state !== 'closed') {
      this.audioContext.close();
    }
  }
};
</script>

<style scoped>
.audio-recorder {
  padding: 20px;
}

.record-button,
.stop-button {
  margin: 5px;
  padding: 10px 20px;
  border-radius: 4px;
  border: none;
  cursor: pointer;
}

.record-button {
  background-color: #ff4444;
  color: white;
}

.record-button:disabled {
  background-color: #ffaaaa;
}

.stop-button {
  background-color: #444444;
  color: white;
}

.stop-button:disabled {
  background-color: #888888;
}

.error-message {
  color: red;
  margin: 10px 0;
}

.duration {
  margin: 10px 0;
  font-size: 1.2em;
}

.playback-control {
  margin: 20px 0;
}

.download-buttons {
  margin-top: 20px;
}

audio {
  width: 100%;
  max-width: 500px;
}
</style>
