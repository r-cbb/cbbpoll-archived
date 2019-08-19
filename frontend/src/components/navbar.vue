<template>
  <div>
    <b-navbar type="is-dark" fixed-top>
      <template slot="brand">
        <b-navbar-item href="/">
          <div class="navbar-item green-text">/r/CollegeBasketball Poll</div>
        </b-navbar-item>
      </template>
      <template slot="start">
        <!-- <b-navbar-item tag="router-link" to="/results">
         Results
        </b-navbar-item> -->
        <b-dropdown hoverable aria-role="list">
          <a id="navDropdown" class="navbar-item" slot="trigger">
            <span>Results</span>
          </a>

          <b-dropdown-item aria-role="listitem" has-link>
            <router-link to="/results">Current Week</router-link>
          </b-dropdown-item>
          <b-dropdown-item aria-role="listitem" has-link>
            <router-link to="/overview">Season</router-link>
          </b-dropdown-item>
        </b-dropdown>

        <b-navbar-item tag="router-link" to="/voters">
          Voters
        </b-navbar-item>
        <b-navbar-item tag="router-link" to="/teams">
          Teams
        </b-navbar-item>
        <b-navbar-item tag="router-link" to="/about">
          About
        </b-navbar-item>
      </template>

      <template slot="end">
        <b-navbar-item>
          <div class="green-text">Login/Sign Up</div>
        </b-navbar-item>
      </template>
    </b-navbar>
  </div>
</template>

<script>
export default {
  name: 'navbar'
}

</script>

<style scoped lang="scss">
.green-text {
  color: #42b983;
}

#navDropdown {
  color: whitesmoke !important;
}
#navDropdown:hover {
  background-color: #292929 !important
}
</style>
