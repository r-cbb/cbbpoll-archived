import Vue from 'vue'
import Router from 'vue-router'
import Home from './views/Home.vue'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/about',
      name: 'about',
      // route level code-splitting
      // this generates a separate chunk (about.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: function () { 
        return import(/* webpackChunkName: "about" */ './views/About.vue')
      }
    },
    {
      path: '/voters',
      name: 'voters',
      component: function () { 
        return import('./views/Voters.vue')
      }
    },
    {
      path: '/teams',
      name: 'teams',
      component: function () { 
        return import('./views/Teams.vue')
      }
    },
    {
      path: '/results',
      name: 'reults',
      component: function () { 
        return import('./views/Results.vue')
      }
    },
    {
      path: '/overview',
      name: 'overview',
      component: function () { 
        return import('./views/Overview.vue')
      }
    }
  ]
})
