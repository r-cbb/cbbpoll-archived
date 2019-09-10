<template>
	<div class="container">
		<section class="section">
			<b-tabs position="is-centered" class="block" type="is-toggle-rounded">
				<b-tab-item label="User">
					user info
				</b-tab-item>
				<b-tab-item label="Team">
					team info
				</b-tab-item>
				<b-tab-item label="Poll">
					poll info
				</b-tab-item>
				<b-tab-item label="Ballot">
					ballot info
				</b-tab-item>
				<b-tab-item label="Vote">
					vote info
				</b-tab-item>
				<b-tab-item label="Vote Event">
					vote event info
				</b-tab-item>
				<b-tab-item label="Voter Application">
					voter application info
				</b-tab-item>
				<b-tab-item label="Consumption Tag">
					consumption tag info
				</b-tab-item>
			</b-tabs>
		</section>
	</div>
</template>

<script>
export default {
	name: 'admin'
}

</script>
