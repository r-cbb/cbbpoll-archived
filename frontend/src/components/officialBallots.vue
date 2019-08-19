<template>
  <div class="has-text-left">
    <div class="title is-6">Official Ballots</div>
    <div ref="ballotWrap">
      <span v-for="user in dummyData" :key="user.username">{{user.username}}, </span>
    </div>
  </div>
</template>

<script>
export default {
  name: 'officialBallots',
  data() {
    return {
      dummyData: [
      {
        username: "testUser",
        flair: "someUrl?"
      },
      {
        username: "basketball",
        flair: ""
      },
      {
        username: "DukeSux",
        flair: ""
      }]
    }
  },
  mounted() {
    let lastBallot = this.$refs.ballotWrap.lastChild.innerHTML
    lastBallot = lastBallot.slice(0, -2)
    this.$refs.ballotWrap.lastChild.innerHTML = lastBallot;
  }
}
</script>

<style scoped lang="scss">
// some styles
</style>
