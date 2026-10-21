<script>
import UserAvatar from './UserAvatar.vue'
import {formatDateTime} from '../services/format.js'

export default {
	components: {UserAvatar},
	props: {
		conversations: {type: Array, required: true},
		selectedId: {type: Number, default: null},
		myId: {type: Number, required: true},
		loaded: {type: Boolean, default: false},
	},
	data() {
		return {
			filter: '',
		}
	},
	computed: {
		filteredConversations() {
			const filter = this.filter.trim().toLowerCase();
			if (filter === '') {
				return this.conversations;
			}
			return this.conversations.filter((c) => c.displayName.toLowerCase().includes(filter));
		},
	},
	methods: {
		formatDateTime,
		senderPrefix(conversation) {
			const message = conversation.lastMessage;
			if (message.sender.id === this.myId) {
				return "Tu: ";
			}
			if (conversation.type === 'group') {
				return `${message.sender.username}: `;
			}
			return "";
		},
		previewText(message) {
			if (message.text) {
				return message.text;
			}
			if (message.type === 'photo') {
				return "Foto";
			}
			if (message.type === 'file') {
				return "File";
			}
			return "";
		},
	},
}
</script>

<template>
	<div class="d-flex flex-column">
		<div class="p-2 border-bottom">
			<input v-model="filter" type="search" class="form-control form-control-sm" placeholder="Cerca chat">
		</div>

		<div class="chat-list-items flex-grow-1 overflow-auto">
			<p v-if="loaded && conversations.length === 0" class="text-center text-body-secondary p-4 mb-0">
				Nessuna conversazione. Inizia una nuova chat!
			</p>
			<p v-else-if="loaded && filteredConversations.length === 0" class="text-center text-body-secondary p-4 mb-0">
				Nessuna chat trovata.
			</p>

			<RouterLink v-for="c in filteredConversations" :key="c.id" :to="`/conversations/${c.id}`"
				class="chat-item d-flex align-items-center gap-2 px-3 py-2 text-decoration-none text-body"
				:class="{selected: c.id === selectedId}">
				<UserAvatar :picture="c.displayPicture" :name="c.displayName" :size="48" />
				<div class="chat-item-body flex-grow-1">
					<div class="d-flex justify-content-between align-items-baseline gap-2">
						<span class="fw-semibold text-truncate">{{ c.displayName }}</span>
						<small class="text-body-secondary text-nowrap">{{ formatDateTime(c.lastMessage.timestamp) }}</small>
					</div>
					<div class="small text-body-secondary text-truncate">
						<i v-if="c.lastMessage.deleted">Messaggio eliminato</i>
						<span v-else-if="c.lastMessage.type === 'system'">{{ c.lastMessage.text }}</span>
						<span v-else>
							{{ senderPrefix(c) }}
							<i v-if="c.lastMessage.type === 'photo'" class="bi bi-image"></i>
							<i v-else-if="c.lastMessage.type === 'file'" class="bi bi-paperclip"></i>
							{{ previewText(c.lastMessage) }}
						</span>
					</div>
				</div>
			</RouterLink>
		</div>
	</div>
</template>

<style>
.chat-list-items {
	min-height: 0;
}

.chat-item:hover {
	background-color: var(--bs-tertiary-bg);
}

.chat-item.selected {
	background-color: var(--fc-accent-soft);
}

.chat-item-body {
	min-width: 0;
}
</style>
