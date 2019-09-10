<template>
  <div class="container">
  	<section class="section">
	  	<div class="title is-4">About The /r/CollegeBasketball User Poll</div>
	  	<hr>
	  	<div class="has-text-left">
		  	<div class="subtitle is-6">This website hosts the /r/CollegeBasketball user poll. This is the third year of the poll, but the first on this website. This poll indicates the relative rankings of the top 25 NCAA Men's Baskeball teams as viewed by a representative sample of the <a href="https://www.reddit.com/r/collegebasketball">/r/CollegeBasketball</a> population. The scoring is typical for polls of this sort. A panel of approximately 65 individuals from around the country (and possibly elsewhere) vote on the poll weekly.</div>
		  	<hr>
		  	<div class="title is-5">Voting Guidelines</div>
		  	<div class="subtitle is-6">The following are the voting and ethics guidelines given to voters:
		  		<ul class="left-margin">
		  			<li>Rank the teams according to <b>your honest perception</b> of their strength, based on personnel, consistent performance, perceived talent/ability (the "eye test"), and coaching.</li>
		  			<li>Teams on NCAA probation <b>are</b> eligible for the poll.</li>
		  			<li>Emphasize performance over achievement.</li>
		  			<li><b>Do not</b> overvalue head-to-head results. The better team does not win every game. (If they did, no teams would split a season series). It's acceptable to rank a team above a team it has lost to immediately. By the end of the season it will be unavoidable.</li>
		  			<li>Consider <b>location</b>. A close road loss can be just as indicative of team strength as a close home win.</li>
		  			<li>Avoid regional/conference bias, for or against. Your team and conference do not deserve any special treatement</li>
		  			<li>Blatant homerism will be challenged, and may lead to dismissal from the poll.</li>
		  		</ul>
		  	</div>
		  	<hr>
		  	<div class="title is-5 margin-bot-5">Frequently Asked Questions and Other Matters of Philosophy</div>
		  	<div class="title is-6">Who makes up the pool of voters?</div>
		  	<div class="subtitle is-6 left-margin">The voters are reddit users active on <a href="https://www.reddit.com/r/collegebasketball">/r/CollegeBasketball</a>. The voters are selected by the subreddit's moderator team in order to keep a reasonable team/conference distribution.</div>
		  	<div class="title is-6">How do I become a voter?</div>
		  	<div class="subtitle is-6 left-margin">Log in to this site with your reddit credentials. Each week, submit a provisional ballot. This ballot will not count toward the overall rankings, but it will demonstrate interest in becoming a voter.<br><br>Occasionally, inactive voters will be removed from the poll and replaced with provisional voters.</div>
		  	<div class="title is-6">I have an algorithm for ranking teams that I would like to use as my ballot, is that ok?</div>
		  	<div class="subtitle is-6 left-margin">Not really. This is a human poll. It is meant to work by aggregating the beliefs of a number of individuals, each with their own unique perspective. You may use your algorithm/statistics as a tool to inform your rankings, but your rankings will ultimately be your rankings. "My computer did it" will not be a good excuse when your rankings make no sense.</div>
		  	<div class="title is-6">How should teams be ranked preseason/before there is much useful information to work from?</div>
		  	<div class="subtitle is-6 left-margin">Use all resources available to you when ranking teams. Consider returning players, as well as the quality of the incoming class. Feel free to refer to third party rankings of recruiting classes. Consider players you believe are primed to have a breakout season. Consider coaches who can consistently coach their teams to success. Be willing to throw out your preseason rankings entirely after a few weeks.</div>
		  	<div class="title is-6">Should teams always drop in my rankings after a loss?</div>
		  	<div class="subtitle is-6 left-margin">If the loss makes you think the team isn't as good as you had previously considered, the team should probably drop in your rankings. If the #10 team lost on the road against the #5 team, there is no reason to think that team is any worse than previously considered. Other polls don't always adhere to common sense--there is no reason you can't. You have our permission to raise the #20 team after a close loss to the #2 team.</div>
		  	<div class="title is-6">Should other polls (Associated Press, USA Today Coaches' Poll, etc) be considered when ranking teams?</div>
		  	<div class="subtitle is-6 left-margin">No.</div>
		  	<div class="title is-6">If I take away one thing from this page, what should it be?</div>
		  	<div class="subtitle is-6 left-margin">A general rule of thumb when ranking teams: Ranking team A above team B indicates that you believe that if a game between the two teams was scheduled on a neutral court four days from today, team A would have a greater than 50% chance of winning. The converse is true as well--ranking a team below another indicates you think they would be neutral court underdogs.</div>
		  </div>
		  <br>
		  <b-message type="is-info">
	      I have a feeling if we could get 50 hoops experts together each week during the season and poll them on who they thought the best team was, we’d get a better power ranking than any computer could produce. If we could free them from the shackles of modern-day voting tendencies, the wisdom of the crowd would provide some very useful information.
	      <br>
	      - Ken Pomeroy in "<a href="https://kenpom.com/blog/">The pre-season AP poll is great</a>""
	    </b-message>

	  </section>
  </div>
</template>

<script>
	export default {
  name: 'about'
}
</script>

<style>
.left-margin {
	margin-left: 2em;
	list-style-type: circle;
	list-style-position: outside;
}
.margin-bot-5 {
	margin-bottom: 5px !important;
}
</style>
