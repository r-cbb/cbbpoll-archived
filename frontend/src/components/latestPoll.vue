<template>
	<div>
		<div class="title is-5">Latest Poll Results</div>
		<div class="subtitle is-7">{{closedTime}}</div>
		<b-table :data="dummyResults" striped>
			<template slot-scope="props">
				<b-table-column label="Rank" width="40" numeric>
					{{ props.row.rank }}
				</b-table-column>
				<b-table-column label="Team (First Place Votes)">
					{{ props.row.team }}
				</b-table-column>
				<b-table-column label="Score">
					{{ props.row.score }}
				</b-table-column>
			</template>
			<template slot="footer">
				<div class="has-text-left" ref="othersWrap">
					<span class="title is-7">Others Receiving Votes: </span>
					<span class="subtitle is-7" v-for="item in othersReceivingVotes" :key="item.team">{{item.team}}({{item.votes}}), </span>
				</div>
			</template>
		</b-table>
	</div>
</template>

<script>
export default {
	name: 'latestPoll',
	data() {
		const dummyResults = [
			{ 'rank': 1, 'team': 'Virginia', 'score': '1500' },
			{ 'rank': 2, 'team': 'Texas Tech', 'score': '1482' },
			{ 'rank': 3, 'team': 'Michigan State', 'score': '1394' },
			{ 'rank': 4, 'team': 'North Carolina', 'score': '1227' },
			{ 'rank': 5, 'team': 'Auburn', 'score': '1226' }
		]

		return {
			dummyResults,
			othersReceivingVotes: [{
					team: 'Iowa State',
					votes: 81
				},
				{
					team: 'Iowa',
					votes: 78
				},
				{
					team: 'Cincinnati',
					votes: 49
				},
				{
					team: 'Nevada',
					votes: 40
				}
			],
			closedTime: "Closed on Friday, April 12, 2019 at 10:00AM EDT"
		}
	},
	mounted() {
		let lastOther = this.$refs.othersWrap.lastChild.innerHTML
		lastOther = lastOther.slice(0, -2)
		this.$refs.othersWrap.lastChild.innerHTML = lastOther;
	}
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">


</style>
