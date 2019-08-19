<template>
  <div class="container">
  	<section class="section">
	  	<div class="title is-4">All 351 Division 1 Teams</div>
	  	<b-table :data="dummyTeams" striped>
				<template slot-scope="props">
					<b-table-column label="School Name">
						{{ props.row.name }}
					</b-table-column>
					<b-table-column label="Logo">
						{{ props.row.logo }}
					</b-table-column>
					<b-table-column label="Nickname">
						{{ props.row.nickname }}
					</b-table-column>
					<b-table-column label="Conference">
						{{ props.row.conference }}
					</b-table-column>
					<b-table-column label="Short Name">
						{{ props.row.short_name }}
					</b-table-column>
					<b-table-column label="Flair Text">
						{{ props.row.flair_text }}
					</b-table-column>
				</template>
			</b-table>
		</section>
  </div>
</template>

<script>
export default {
  data() {
		return {
			dummyTeams: [
				{ 'name': 'Abilene Christian University', 'logo': 'someUrl', 'nickname': 'Wildcats', 'conference': 'Southland', 'short_name': 'Abilene Christian', 'flair_text': 'abilenechristian' },
				{ 'name': 'United States Air Force Academy', 'logo': 'someUrl', 'nickname': 'Falcons', 'conference': 'Mountain West', 'short_name': 'Air Force', 'flair_text': 'airforce' },
				{ 'name': 'University of Akron', 'logo': 'someUrl', 'nickname': 'Zips', 'conference': 'MAC', 'short_name': 'Akron', 'flair_text': 'akron' },
				{ 'name': 'University of Alabama	', 'logo': 'someUrl', 'nickname': 'Crimson Tide', 'conference': 'SEC', 'short_name': 'Alabama', 'flair_text': 'alabama' },
				{ 'name': 'Alabama Agricultural and Mechanical University', 'logo': 'someUrl', 'nickname': 'Bulldogs', 'conference': 'SWAC', 'short_name': 'Alabama A&M', 'flair_text': 'alabamaam' }
			]
		}
	}
}
</script>
