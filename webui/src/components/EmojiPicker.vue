<script>
const reactionEmojis = ["👍", "❤️", "😂", "😮", "😢", "🙏"];

export default {
	props: {
		selected: {type: String, default: null},
	},
	emits: ['select', 'close'],
	data() {
		return {
			emojis: reactionEmojis,
		}
	},
}
</script>

<template>
	<div class="emoji-picker d-flex align-items-center gap-1 p-1 bg-body border rounded-pill shadow">
		<button v-for="emoji in emojis" :key="emoji" type="button" class="btn btn-sm emoji-option rounded-circle"
			:class="{selected: emoji === selected}" :title="emoji === selected ? 'Togli la reazione' : 'Reagisci'"
			@click="$emit('select', emoji)">
			{{ emoji }}
		</button>
		<button type="button" class="btn btn-sm rounded-circle" title="Chiudi" @click="$emit('close')">
			<i class="bi bi-x-lg"></i>
		</button>
	</div>
</template>

<style>
.emoji-option {
	font-size: 1.25rem;
	line-height: 1;
	padding: 0.35rem;
}

.emoji-option.selected {
	background-color: var(--fc-accent-soft);
	outline: 2px solid var(--fc-accent);
}
</style>
