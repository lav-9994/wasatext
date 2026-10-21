<script>
const avatarColors = ["#e57373", "#f06292", "#ba68c8", "#7986cb", "#4fc3f7", "#4db6ac", "#81c784", "#ffb74d", "#a1887f", "#90a4ae"];

export default {
	props: {
		picture: {type: String, default: null},
		name: {type: String, default: ""},
		size: {type: Number, default: 40},
	},
	computed: {
		initial() {
			const first = [...(this.name || "").trim()][0];
			return first ? first.toUpperCase() : "?";
		},
		color() {
			let hash = 0;
			for (const character of this.name || "") {
				hash = (Math.imul(hash, 16777619) ^ character.codePointAt(0)) >>> 0;
			}
			return avatarColors[hash % avatarColors.length];
		},
		boxStyle() {
			return {width: `${this.size}px`, height: `${this.size}px`, minWidth: `${this.size}px`};
		},
		initialStyle() {
			return {...this.boxStyle, backgroundColor: this.color, fontSize: `${Math.round(this.size * 0.45)}px`};
		},
	},
}
</script>

<template>
	<img v-if="picture" :src="picture" :alt="name" class="rounded-circle object-fit-cover" :style="boxStyle">
	<div v-else class="rounded-circle d-flex align-items-center justify-content-center text-white fw-semibold"
		:style="initialStyle">
		{{ initial }}
	</div>
</template>

<style>
</style>
