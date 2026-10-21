<script>
import EmojiPicker from './EmojiPicker.vue'
import {formatDateTime} from '../services/format.js'
import {attachmentURL} from '../services/files.js'

export default {
	components: {EmojiPicker},
	props: {
		message: {type: Object, required: true},
		quoted: {type: Object, default: null},
		myId: {type: Number, required: true},
	},
	emits: ['reply', 'forward', 'react', 'unreact', 'delete', 'open-photo', 'open-user'],
	data() {
		return {
			menuOpen: false,
			pickerOpen: false,
		}
	},
	computed: {
		mine() {
			return this.message.sender.id === this.myId;
		},
		canAct() {
			return this.message.type !== 'system' && !this.message.deleted;
		},
		time() {
			return formatDateTime(this.message.timestamp);
		},
		fileURL() {
			return this.message.file ? attachmentURL(this.message.file) : null;
		},
		myReaction() {
			const comment = this.message.comments.find((c) => c.user.id === this.myId);
			return comment ? comment.emoji : null;
		},
		reactionGroups() {
			const groups = [];
			for (const comment of this.message.comments) {
				let group = groups.find((g) => g.emoji === comment.emoji);
				if (!group) {
					group = {emoji: comment.emoji, names: []};
					groups.push(group);
				}
				group.names.push(comment.user.id === this.myId ? "Tu" : comment.user.username);
			}
			return groups;
		},
		quotedSender() {
			if (!this.quoted) {
				return "";
			}
			return this.quoted.sender.id === this.myId ? "Tu" : this.quoted.sender.username;
		},
	},
	methods: {
		closeAll() {
			this.menuOpen = false;
			this.pickerOpen = false;
		},
		toggleMenu() {
			this.pickerOpen = false;
			this.menuOpen = !this.menuOpen;
		},
		choose(action) {
			this.menuOpen = false;
			if (action === 'react') {
				this.pickerOpen = true;
			} else {
				this.$emit(action, this.message);
			}
		},
		selectEmoji(emoji) {
			this.pickerOpen = false;
			if (emoji === this.myReaction) {
				this.$emit('unreact', this.message);
			} else {
				this.$emit('react', this.message, emoji);
			}
		},
	},
}
</script>

<template>
	<div v-if="message.type === 'system'" class="text-center my-2">
		<span class="system-message d-inline-block small text-body-secondary px-3 py-1 rounded-pill">
			{{ message.text }} · {{ time }}
		</span>
	</div>

	<div v-else class="d-flex mb-2" :class="mine ? 'justify-content-end' : 'justify-content-start'">
		<div class="message-wrapper position-relative" :class="mine ? 'align-items-end' : 'align-items-start'">
			<div class="bubble shadow-sm" :class="mine ? 'bubble-mine' : 'bubble-other'">
				<div class="d-flex align-items-start gap-1">
					<div class="bubble-content flex-grow-1">
						<button v-if="!mine" type="button"
							class="sender-name d-block border-0 bg-transparent p-0 small fw-semibold text-start"
							title="Vedi profilo" @click="$emit('open-user', message.sender.id)">
							{{ message.sender.username }}
						</button>

						<div v-if="message.forwarded && !message.deleted" class="small fst-italic text-body-secondary">
							<i class="bi bi-forward-fill"></i> Inoltrato
						</div>

						<div v-if="message.inReplyTo && !message.deleted" class="reply-quote small mb-1">
							<template v-if="quoted">
								<div class="fw-semibold">{{ quotedSender }}</div>
								<div class="text-truncate text-body-secondary">
									<i v-if="quoted.deleted">Messaggio eliminato</i>
									<template v-else>
										<i v-if="quoted.type === 'photo'" class="bi bi-image"></i>
										<i v-else-if="quoted.type === 'file'" class="bi bi-paperclip"></i>
										{{ quoted.text || (quoted.type === 'photo' ? 'Foto' : quoted.type === 'file' ? quoted.file.name : '') }}
									</template>
								</div>
							</template>
							<i v-else class="text-body-secondary">Messaggio non disponibile</i>
						</div>

						<i v-if="message.deleted" class="text-body-secondary">Messaggio eliminato</i>
						<template v-else>
							<img v-if="message.photo" :src="message.photo" alt="Foto" class="message-photo rounded mb-1"
								@click="$emit('open-photo', message.photo)">
							<a v-if="message.file" :href="fileURL" :download="message.file.name"
								class="file-link d-flex align-items-center gap-2 rounded p-2 mb-1 text-decoration-none text-body">
								<i class="bi bi-file-earmark-arrow-down fs-4"></i>
								<span class="text-truncate">{{ message.file.name }}</span>
							</a>
							<div v-if="message.text" class="message-text">{{ message.text }}</div>
						</template>

						<div class="message-meta text-end text-body-secondary">
							{{ time }}
							<template v-if="mine && message.status">
								<i v-if="message.status === 'sent'" class="bi bi-clock ms-1" title="Inviato"></i>
								<i v-else-if="message.status === 'received'" class="bi bi-check-lg ms-1" title="Ricevuto"></i>
								<i v-else class="bi bi-check-all ms-1 status-read" title="Letto"></i>
							</template>
						</div>
					</div>

					<button v-if="canAct" type="button" class="btn btn-sm menu-button text-body-secondary" title="Azioni"
						@click="toggleMenu">
						<i class="bi bi-three-dots"></i>
					</button>
				</div>
			</div>

			<div v-if="reactionGroups.length > 0" class="d-flex flex-wrap gap-1 mt-1"
				:class="mine ? 'justify-content-end' : 'justify-content-start'">
				<span v-for="group in reactionGroups" :key="group.emoji"
					class="reaction small bg-body border rounded-pill px-2 py-1">
					{{ group.emoji }} {{ group.names.join(', ') }}
				</span>
			</div>

			<div v-if="menuOpen || pickerOpen" class="message-backdrop" @click="closeAll"></div>

			<div v-if="menuOpen" class="message-menu list-group shadow" :class="mine ? 'end-0' : 'start-0'">
				<button type="button" class="list-group-item list-group-item-action" @click="choose('reply')">
					<i class="bi bi-reply me-2"></i>Rispondi
				</button>
				<button type="button" class="list-group-item list-group-item-action" @click="choose('forward')">
					<i class="bi bi-forward me-2"></i>Inoltra
				</button>
				<button type="button" class="list-group-item list-group-item-action" @click="choose('react')">
					<i class="bi bi-emoji-smile me-2"></i>Reagisci
				</button>
				<button v-if="mine" type="button" class="list-group-item list-group-item-action text-danger"
					@click="choose('delete')">
					<i class="bi bi-trash me-2"></i>Elimina
				</button>
			</div>

			<div v-if="pickerOpen" class="message-picker" :class="mine ? 'end-0' : 'start-0'">
				<EmojiPicker :selected="myReaction" @select="selectEmoji" @close="pickerOpen = false" />
			</div>
		</div>
	</div>
</template>

<style>
.message-wrapper {
	display: flex;
	flex-direction: column;
	max-width: min(75%, 560px);
}

.bubble {
	padding: 0.4rem 0.6rem;
	border-radius: 0.9rem;
	overflow-wrap: anywhere;
}

.bubble-mine {
	background-color: var(--fc-accent-soft);
	border-bottom-right-radius: 0.25rem;
}

.bubble-other {
	background-color: var(--bs-body-bg);
	border-bottom-left-radius: 0.25rem;
}

.bubble-content {
	min-width: 0;
}

.sender-name {
	color: var(--fc-accent);
}

.reply-quote {
	border-left: 3px solid var(--fc-accent);
	background-color: var(--bs-tertiary-bg);
	border-radius: 0.25rem;
	padding: 0.25rem 0.5rem;
}

.message-text {
	white-space: pre-wrap;
}

.message-photo {
	display: block;
	max-width: 100%;
	max-height: 320px;
	cursor: zoom-in;
}

.file-link {
	background-color: var(--bs-tertiary-bg);
}

.message-meta {
	font-size: 0.75rem;
}

.status-read {
	color: var(--fc-accent);
}

.menu-button {
	padding: 0 0.25rem;
	line-height: 1;
}

.system-message {
	background-color: var(--bs-secondary-bg);
	max-width: 85%;
}

.message-backdrop {
	position: fixed;
	inset: 0;
	z-index: 20;
}

.message-menu,
.message-picker {
	position: absolute;
	top: 100%;
	z-index: 21;
	margin-top: 0.25rem;
}

.message-menu {
	min-width: 180px;
}
</style>
