<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :span="6">
        <div class="stat-card">
          <div class="icon icon-gradient-primary">
            <i class="fa fa-shield"></i>
          </div>
          <div class="info">
            <div class="title">WebSocket 状态</div>
            <div class="value" :class="webSocketStatus ? 'status-connected' : 'status-disconnected'">
              {{ webSocketStatus ? '已连接' : '已断开' }}
            </div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="icon icon-gradient-warning">
            <i class="fa fa-bell"></i>
          </div>
          <div class="info">
            <div class="title">报警次数</div>
            <div class="value">{{ warningCount }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="icon icon-gradient-success">
            <i class="fa fa-clock-o"></i>
          </div>
          <div class="info">
            <div class="title">运行时间</div>
            <div class="value">{{ allTime || '0:00:00' }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="icon icon-gradient-info">
            <i class="fa fa-plug"></i>
          </div>
          <div class="info">
            <div class="title">已加载插件</div>
            <div class="value">{{ plugData.length }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 主要内容区 -->
    <el-row :gutter="20" class="content-row">
      <el-col :span="16">
        <div class="card">
          <div class="card-header">
            <h3>系统警告</h3>
            <el-button type="text" @click="showWarnLog()">查看详情 <i class="fa fa-angle-right"></i></el-button>
          </div>
          <el-table 
            :data="tableData" 
            style="width: 100%"
            :header-cell-style="{ background: '#f5f7fa', color: '#606266' }">
            <el-table-column prop="time" label="时间" width="180">
              <template slot-scope="scope">
                <span class="time-cell">{{ scope.row.time }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="120">
              <template slot-scope="scope">
                <el-tag :type="getTagType(scope.row.type)" size="small">{{ scope.row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="plugin" label="插件" width="120"></el-table-column>
            <el-table-column prop="message" label="描述" min-width="45"></el-table-column>
          </el-table>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="card">
          <div class="card-header">
            <h3>已载入插件</h3>
            <el-button type="text" @click="reloadPlugin()">重载 <i class="fa fa-refresh"></i></el-button>
          </div>
          <div class="plugin-list">
            <div v-for="(plugin, index) in plugData" :key="index" class="plugin-item">
              <div class="plugin-icon">
                <i class="fa fa-puzzle-piece"></i>
              </div>
              <div class="plugin-name">{{ plugin.name }}</div>
            </div>
            <div v-if="plugData.length === 0" class="empty-state">
              <i class="fa fa-inbox"></i>
              <p>暂无插件</p>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 更新时间 -->
    <div class="update-time">
      <i class="fa fa-refresh"></i> 最后更新: {{ lastUpdateTime || '暂无数据' }}
    </div>
  </div>
</template>

<script>
import config from "../config.js";
import axios from "axios";
import { mapGetters } from "vuex";

export default {
  mounted() {
    this.refresh();
    this.getPlugin();
    this.$bus.on("goto-main-latest", () => {
      this.refresh();
    });
  },
  computed: {
    ...mapGetters({
      webSocketStatus: "getWsState"
    })
  },
  data() {
    return {
      lastUpdateTime: "",
      warningCount: 0,
      allTime: "",
      tableData: [],
      plugData: []
    };
  },
  methods: {
    refresh() {
      axios
        .get(config.ajax_addr + `/listalert?page=0&count=8`)
        .then(res => {
          this.tableData = res.data.data || [];
        });
      axios.get(config.ajax_addr + "/info").then(res => {
        this.lastUpdateTime = res.data.timestamp_lastupdate;
        this.warningCount = res.data.count_alert;
        this.allTime = res.data.timestamp_runningtime;
      });
    },

    showWarnLog() {
      this.$router.push({
        path: "../warnLog"
      });
    },

    reloadPlugin() {
      this.$message.info('正在重载插件...');
      axios.get(`${config.ajax_addr}/reloadplugin`).then(res => {
        this.getPlugin();
        this.$message.success('插件重载成功');
      }).catch(err => {
        this.$message.error('插件重载失败');
      });
    },

    getPlugin() {
      axios.get(`${config.ajax_addr}/listplugin`).then(res => {
        this.plugData = [];
        res.data.data.forEach(ele => {
          this.plugData.push({
            name: ele
          });
        });
      });
    },

    getTagType(type) {
      const types = {
        'Web': 'warning',
        'PWN': 'danger',
        'File': 'success',
        'Process': 'info'
      };
      return types[type] || '';
    }
  }
};
</script>

<style lang="scss" scoped>
.dashboard {
  padding: 10px;
}

.stat-row {
  margin-bottom: 20px;
  
  .stat-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    display: flex;
    align-items: center;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
    
    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
    }
    
    .icon {
      width: 56px;
      height: 56px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 16px;
      
      i {
        font-size: 28px;
        color: #fff;
      }
    }
    
    .info {
      flex: 1;
      
      .title {
        font-size: 14px;
        color: #909399;
        margin-bottom: 8px;
      }
      
      .value {
        font-size: 28px;
        font-weight: 600;
        color: #303133;
      }
      
      .status-connected {
        color: #67C23A;
      }
      
      .status-disconnected {
        color: #F56C6C;
      }
    }
  }
}

.content-row {
  margin-bottom: 20px;
}

.card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 20px;
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 15px;
    border-bottom: 1px solid #ebeef5;
    
    h3 {
      margin: 0;
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }
}

.time-cell {
  font-family: monospace;
  font-size: 13px;
  color: #606266;
}

.plugin-list {
  max-height: 300px;
  overflow-y: auto;
}

.plugin-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 8px;
  background: #f5f7fa;
  transition: all 0.3s ease;
  
  &:hover {
    background: #ecf5ff;
  }
  
  &:last-child {
    margin-bottom: 0;
  }
  
  .plugin-icon {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 12px;
    
    i {
      color: #fff;
      font-size: 16px;
    }
  }
  
  .plugin-name {
    font-size: 14px;
    color: #303133;
    font-weight: 500;
  }
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #c0c4cc;
  
  i {
    font-size: 48px;
    margin-bottom: 16px;
    display: block;
  }
  
  p {
    margin: 0;
    font-size: 14px;
  }
}

.update-time {
  text-align: right;
  font-size: 12px;
  color: #909399;
  
  i {
    margin-right: 4px;
  }
}
</style>
