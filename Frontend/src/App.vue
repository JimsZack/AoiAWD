<template>
  <div id="app">
    <transition name="fade" mode="out-in">
      <router-view></router-view>
    </transition>
  </div>
</template>

<script>
import config from "./config.js";

const RECONNECT_MIN = 1000;
const RECONNECT_MAX = 10000;

export default {
  name: "app",
  data() {
    return {
      ws: null,
      reconnectTimer: null,
      reconnectDelay: RECONNECT_MIN,
      destroyed: false
    };
  },
  mounted: function() {
    this.showWSInput();
  },
  beforeDestroy: function() {
    this.destroyed = true;
    this.clearReconnectTimer();
    this.closeWS();
  },
  components: {},
  methods: {
    showWSInput() {
      if (this.ws) {
        return;
      }
      var ws;
      try {
        ws = new WebSocket(config.ws_addr);
      } catch (e) {
        this.scheduleReconnect();
        return;
      }
      this.ws = ws;
      ws.onopen = () => {
        this.reconnectDelay = RECONNECT_MIN;
        this.$store.dispatch("openWs");
      };
      ws.onclose = () => {
        this.ws = null;
        this.$store.dispatch("closeWs");
        if (this.destroyed) {
          return;
        }
        this.$message.error("WebSocket连接丢失，正在尝试重连");
        this.scheduleReconnect();
      };
      ws.onmessage = msg => {
        this.dispatchMessage(msg);
      };
    },
    dispatchMessage(msg) {
      var data;
      try {
        data = JSON.parse(msg.data);
      } catch (e) {
        return;
      }
      var type = data && data.type;
      this.$bus.emit("goto-main-latest");
      switch (type) {
        case "file":
          this.$bus.emit("goto-file-latest");
          break;
        case "process":
          this.$bus.emit("goto-process-latest");
          this.$bus.emit("process-working");
          break;
        case "web":
          this.$bus.emit("goto-web-latest");
          break;
        case "alert":
          this.$bus.emit("goto-alert-latest");
          break;
        case "pwn":
          this.$bus.emit("goto-pwn-latest");
          break;
      }
    },
    scheduleReconnect() {
      if (this.destroyed || this.reconnectTimer) {
        return;
      }
      var delay = this.reconnectDelay;
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, RECONNECT_MAX);
      var _this = this;
      this.reconnectTimer = setTimeout(() => {
        _this.reconnectTimer = null;
        _this.showWSInput();
      }, delay);
    },
    clearReconnectTimer() {
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
    },
    closeWS() {
      var ws = this.ws;
      this.ws = null;
      if (!ws) {
        return;
      }
      ws.onopen = null;
      ws.onmessage = null;
      ws.onclose = null;
      try {
        ws.close();
      } catch (e) {
        // 忽略关闭异常
      }
    }
  }
};
</script>

<style lang="scss">
body {
  margin: 0px;
  padding: 0px;
  font-family: Helvetica Neue, Helvetica, PingFang SC, Hiragino Sans GB,
    Microsoft YaHei, SimSun, sans-serif;
  font-size: 14px;
  -webkit-font-smoothing: antialiased;
}

#app {
  position: absolute;
  top: 0px;
  bottom: 0px;
  width: 100%;
}

.el-submenu [class^="fa"] {
  vertical-align: baseline;
  margin-right: 10px;
}

.el-menu-item [class^="fa"] {
  vertical-align: baseline;
  margin-right: 10px;
}

.toolbar {
  background: #f2f2f2;
  padding: 10px;
  //border:1px solid #dfe6ec;
  margin: 10px 0px;
  .el-form-item {
    margin-bottom: 10px;
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s ease;
}

.fade-enter,
.fade-leave-active {
  opacity: 0;
}
</style>