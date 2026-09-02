import Vue from 'vue'
import App from './App'
import ElementUI from 'element-ui'
import 'element-ui/lib/theme-default/index.css'
import VueRouter from 'vue-router'
import store from './vuex/store'
import Vuex from 'vuex'
import VueBus from 'vue-bus';
import routes from './routes'
import 'font-awesome/css/font-awesome.min.css'
import './styles/global.scss'
import Axios from 'axios';


Vue.use(ElementUI)
Vue.use(VueRouter)
Vue.use(Vuex)
Vue.use(VueBus)


const router = new VueRouter({
  routes
})

router.beforeEach((to, from, next) => {
  if (to.path == "/login") {
    next();
    return;
  }
  var token = sessionStorage.getItem('accessToken');
  if (!token) {
    next("/login");
    return;
  }
  Axios.defaults.headers['Token'] = token;
  next();
})

Axios.interceptors.response.use(response => {
  return response
},
  err => {
    var status = err.response && err.response.status;
    if (status === 401 || status === 403) {
      sessionStorage.removeItem('accessToken');
      delete Axios.defaults.headers['Token'];
      if (router.currentRoute.path !== '/login') {
        router.push({
          path: '/login'
        })
      }
    }
    return Promise.reject(err);
  }
)

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')

