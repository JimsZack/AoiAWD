<template>
	<el-row class="container">
		<el-col :span="24" class="header">
			<el-col :span="10" class="logo" :class="collapsed?'logo-collapse-width':'logo-width'">
				<div class="logo-wrapper" v-if="!collapsed">
					<svg viewBox="0 0 120 120" class="logo-svg">
						<defs>
							<linearGradient id="headerGrad" x1="0%" y1="0%" x2="100%" y2="100%">
								<stop offset="0%" style="stop-color:#fff;stop-opacity:1" />
								<stop offset="100%" style="stop-color:#e0e0e0;stop-opacity:1" />
							</linearGradient>
						</defs>
						<circle cx="60" cy="60" r="40" fill="none" stroke="url(#headerGrad)" stroke-width="3" />
						<text x="60" y="72" font-size="32" fill="white" text-anchor="middle" font-weight="bold">G</text>
					</svg>
					<span class="logo-text">{{sysName}}</span>
				</div>
				<div class="logo-wrapper collapsed-logo" v-else>
					<svg viewBox="0 0 120 120" class="logo-svg">
						<circle cx="60" cy="60" r="40" fill="none" stroke="white" stroke-width="3" />
						<text x="60" y="72" font-size="32" fill="white" text-anchor="middle" font-weight="bold">G</text>
					</svg>
				</div>
			</el-col>
			<el-col :span="6">
				<div class="tools" @click.prevent="collapse">
					<i class="fa fa-align-justify"></i>
				</div>
			</el-col>
			<el-col :span="8" class="userinfo">
				<div class="user-info-inner">
					<span class="version-badge">v2.0</span>
					<a href="https://github.com/JimsZack/AoiAWD" target="_blank" class="github-link">
						<i class="fa fa-github"></i>
					</a>
				</div>
			</el-col>
		</el-col>
		<el-col :span="24" class="main">
			<aside :class="collapsed?'menu-collapsed':'menu-expanded'">
				<!--导航菜单-->
				<el-menu :default-active="$route.path" class="el-menu-vertical-demo" @open="handleopen" @close="handleclose" @select="handleselect"
					 unique-opened router v-show="!collapsed">
					<template v-for="(item,index) in $router.options.routes" v-if="!item.hidden">
						<el-submenu :index="index+''" v-if="!item.leaf">
							<template slot="title"><i :class="item.iconCls"></i>{{item.name}}</template>
							<el-menu-item v-for="child in item.children" :index="child.path" :key="child.path" v-if="!child.hidden">{{child.name}}</el-menu-item>
						</el-submenu>
						<el-menu-item v-if="item.leaf&&item.children.length>0" :index="item.children[0].path"><i :class="item.iconCls"></i>{{item.children[0].name}}</el-menu-item>
					</template>
				</el-menu>
				<!--导航菜单-折叠后-->
				<ul class="el-menu el-menu-vertical-demo collapsed" v-show="collapsed" ref="menuCollapsed">
					<li v-for="(item,index) in $router.options.routes" v-if="!item.hidden" class="el-submenu item">
						<template v-if="!item.leaf">
							<div class="el-submenu__title" style="padding-left: 20px;" @mouseover="showMenu(index,true)" @mouseout="showMenu(index,false)"><i :class="item.iconCls"></i></div>
							<ul class="el-menu submenu" :class="'submenu-hook-'+index" @mouseover="showMenu(index,true)" @mouseout="showMenu(index,false)"> 
								<li v-for="child in item.children" v-if="!child.hidden" :key="child.path" class="el-menu-item" style="padding-left: 40px;" :class="$route.path==child.path?'is-active':''" @click="$router.push(child.path)">{{child.name}}</li>
							</ul>
						</template>
						<template v-else>
							<li class="el-submenu">
								<div class="el-submenu__title el-menu-item" style="padding-left: 20px;height: 56px;line-height: 56px;padding: 0 20px;" :class="$route.path==item.children[0].path?'is-active':''" @click="$router.push(item.children[0].path)"><i :class="item.iconCls"></i></div>
							</li>
						</template>
					</li>
				</ul>
			</aside>
			<section class="content-container">
				<div class="grid-content bg-purple-light">
					<el-col :span="24" class="breadcrumb-container">
						<el-breadcrumb separator="/" class="breadcrumb-inner">
							<el-breadcrumb-item v-for="item in $route.matched" :key="item.path">
								{{ item.name }}
							</el-breadcrumb-item>
						</el-breadcrumb>
					</el-col>
					<el-col :span="24" class="content-wrapper" style="margin-top:20px;">
						<transition name="fade" mode="out-in">
							<router-view></router-view>
						</transition>
					</el-col>
				</div>
			</section>
		</el-col>
	</el-row>
</template>

<script>
	export default {
		data() {
			return {
				sysName:'GoAWD',
				collapsed:false,
				sysUserName: '',
				sysUserAvatar: '',
				form: {
					name: '',
					region: '',
					date1: '',
					date2: '',
					delivery: false,
					type: [],
					resource: '',
					desc: ''
				}
			}
		},
		methods: {
			onSubmit() {
				console.log('submit!');
			},
			handleopen() {
				//console.log('handleopen');
			},
			handleclose() {
				//console.log('handleclose');
			},
			handleselect: function (a, b) {
			},
			//退出登录
			logout: function () {
				var _this = this;
				this.$confirm('确认退出吗?', '提示', {
					//type: 'warning'
				}).then(() => {
					sessionStorage.removeItem('user');
					_this.$router.push('/login');
				}).catch(() => {
				});
			},
			//折叠导航栏
			collapse:function(){
				this.collapsed=!this.collapsed;
			},
			showMenu(i,status){
				this.$refs.menuCollapsed.getElementsByClassName('submenu-hook-'+i)[0].style.display=status?'block':'none';
			}
		},
		mounted() {
			var user = sessionStorage.getItem('user');
			if (user) {
				user = JSON.parse(user);
				this.sysUserName = user.name || '';
				this.sysUserAvatar = user.avatar || '';
			}
		}
	}
</script>

<style scoped lang="scss">
	.container {
		position: absolute;
		top: 0px;
		bottom: 0px;
		width: 100%;
		
		.header {
			height: 60px;
			line-height: 60px;
			background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
			color:#fff;
			
			.userinfo {
				text-align: right;
				padding-right: 20px;
				float: right;
				
				.user-info-inner {
					display: flex;
					align-items: center;
					justify-content: flex-end;
					height: 100%;
				}
				
				.version-badge {
					background: rgba(255, 255, 255, 0.2);
					padding: 4px 10px;
					border-radius: 12px;
					font-size: 12px;
					margin-right: 16px;
				}
				
				.github-link {
					color: #fff;
					font-size: 20px;
					transition: all 0.3s;
					
					&:hover {
						color: #667eea;
						transform: scale(1.1);
					}
				}
			}
			
			.logo {
				height:60px;
				font-size: 22px;
				padding-left:20px;
				padding-right:20px;
				border-right-width: 1px;
				border-right-style: solid;
				border-color: rgba(255, 255, 255, 0.1);
				
				.logo-wrapper {
					display: flex;
					align-items: center;
					height: 100%;
					
					.logo-svg {
						width: 36px;
						height: 36px;
						margin-right: 10px;
					}
					
					.logo-text {
						font-weight: 600;
						letter-spacing: 1px;
					}
				}
				
				.collapsed-logo {
					justify-content: center;
				}
			}
			
			.logo-width{
				width:200px;
			}
			.logo-collapse-width{
				width:60px
			}
			
			.tools{
				padding: 0px 23px;
				width:14px;
				height: 60px;
				line-height: 60px;
				cursor: pointer;
				transition: all 0.3s;
				
				&:hover {
					color: #667eea;
				}
			}
		}
		
		.main {
			display: flex;
			position: absolute;
			top: 60px;
			bottom: 0px;
			overflow: hidden;
			
			aside {
				flex:0 0 230px;
				width: 230px;
				background: #252a34;
				
				.el-menu{
					height: 100%;
					border-right: none;
				}
				
				.collapsed{
					width:60px;
					.item{
						position: relative;
					}
					.submenu{
						position:absolute;
						top:0px;
						left:60px;
						z-index:99999;
						height:auto;
						display:none;
					}
				}
			}
			
			.menu-collapsed{
				flex:0 0 60px;
				width: 60px;
			}
			
			.menu-expanded{
				flex:0 0 230px;
				width: 230px;
			}
			
			.content-container {
				flex:1;
				overflow-y: scroll;
				padding: 20px;
				background: #f0f2f5;
				
				.breadcrumb-container {
					.title {
						width: 200px;
						float: left;
						color: #475669;
					}
					.breadcrumb-inner {
						float: right;
					}
				}
				
				.content-wrapper {
					background-color: #fff;
					border-radius: 8px;
					box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
					box-sizing: border-box;
					padding: 20px;
				}
			}
		}
	}
</style>
