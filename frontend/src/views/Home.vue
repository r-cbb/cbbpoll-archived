<template>
  <div>
    <section class="hero is-light">
        <div class="hero-body has-text-left">
          <div class="container">
            <h1 class="title">
              Welcome!
            </h1>
            <h2 class="subtitle">
              This is the site for the <a class="has-text-link" href="https://www.reddit.com/r/CollegeBasketball/">/r/CollegeBasketball</a> User Poll. Applications for voters are currently being accepted. Please log in to apply or submit a ballot.
            </h2>
            <router-link to="/login"><b-button type="is-primary">Login Using Reddit</b-button></router-link>
          </div>
        </div>
      </section>
    <div class="container">
      <section class="section">
      	
        <br>
        <div class="tile is-ancestor">
          <div class="tile is-parent is-vertical is-8">
            <div class="tile is-child box">
              <latest/>
            </div>
          </div>
          <div class="tile is-parent is-vertical is-4">
            <div class="tile is-child box">
              <officialBallots/>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import latest from '@/components/latestPoll.vue'
import officialBallots from '@/components/officialBallots.vue'

export default {
  name: 'home',
  components: {
    latest,
    officialBallots
  }
}
</script>
