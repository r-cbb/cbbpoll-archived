<template>
  <div id="app">
    <navbar/>
    <div id="wrap">
      <router-view />
    </div>
    
  </div>
</template>

<script>
  import navbar from './components/navbar.vue'

  export default {
    name: 'app',
    components: {
      navbar
    },
    data(){
      return {

      }
    },
    methods: {

    }
  }
</script>

<style lang="scss">
@import "~bulma/sass/utilities/_all";

html,
body {
  overflow: auto !important;
  height: 100%;
  width: 100%;
  max-height: 100%;
}

#app {
  font-family: 'Roboto', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

#wrap {
  top: 3.25rem;
  left: 0;
  right: 0;
  bottom: 0;
  position: fixed;
  overflow-y: auto;
}

// Set your colors
$primary: #347AB7;
$primary-invert: findColorInvert($primary);
$twitter: #4099FF;
$twitter-invert: findColorInvert($twitter);

// Setup $colors to use as bulma classes (e.g. 'is-twitter')
$colors: (
    "white": ($white, $black),
    "black": ($black, $white),
    "light": ($light, $light-invert),
    "dark": ($dark, $dark-invert),
    "primary": ($primary, $primary-invert),
    "info": ($info, $info-invert),
    "success": ($success, $success-invert),
    "warning": ($warning, $warning-invert),
    "danger": ($danger, $danger-invert),
    "twitter": ($twitter, $twitter-invert),
    "link": ($link, $link-invert)
);

// Links
$link: $primary;
$link-invert: $primary-invert;
$link-focus-border: $primary;

.navbar {
  // box-shadow: 0px 1px 3px grey;
  height: 3.25rem;
}

// Import Bulma and Buefy styles
@import "~bulma";
@import '~buefy/dist/buefy.css'

</style>
