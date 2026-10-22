<script>
import UserAvatar from './UserAvatar.vue'
import MessageItem from './MessageItem.vue'
import MessageComposer from './MessageComposer.vue'
import ForwardDialog from './ForwardDialog.vue'
import GroupInfoPanel from './GroupInfoPanel.vue'
import UserProfilePanel from './UserProfilePanel.vue'

const refreshInterval = 2500;

export default {
	components: {UserAvatar, MessageItem, MessageComposer, ForwardDialog, GroupInfoPanel, UserProfilePanel},
	props: {
		conversationId: {type: Number, required: true},
		myId: {type: Number, required: true},
	},
	emits: ['changed'],
	data() {
		return {
			conversation: null,
			notFound: false,
			errormsg: null,
			replyTo: null,
			forwarding: null,
			viewedPhoto: null,
			infoPanel: null,
			active: false,
			timer: null,
		}
	},
	computed: {
		messages() {
			return this.conversation ? [...this.conversation.messages].reverse() : [];
		},
		messagesById() {
			const byId = {};
			for (const message of this.messages) {
				byId[message.id] = message;
			}
			return byId;
		},
		otherMemberId() {
			if (!this.conversation || this.conversation.type !== 'direct') {
				return null;
			}
			const other = this.conversation.members.find((m) => m.id !== this.myId);
			return other ? other.id : null;
		},
		membersText() {
			if (!this.conversation || this.conversation.type !== 'group') {
				return "";
			}
			return this.conversation.members.map((m) => m.id === this.myId ? "Tu" : m.username).join(", ");
		},
	},
	methods: {
		isAtBottom() {
			const list = this.$refs.messageList;
			return !list || list.scrollHeight - list.scrollTop - list.clientHeight < 80;
		},
		scrollToBottom() {
			const list = this.$refs.messageList;
			if (list) {
				list.scrollTop = list.scrollHeight;
			}
		},
		async refresh(forceScroll) {
			const firstLoad = this.conversation === null;
			const stickToBottom = forceScroll || firstLoad || this.isAtBottom();
			try {
				const response = await this.$axios.get(`/conversations/${this.conversationId}`);
				this.conversation = response.data;
				this.errormsg = null;
				if (this.replyTo && !this.conversation.messages.some((m) => m.id === this.replyTo.id && !m.deleted)) {
					this.replyTo = null;
				}
				if (stickToBottom) {
					this.$nextTick(this.scrollToBottom);
				}
			} catch (e) {
				if (e.response && e.response.status === 404) {
					this.notFound = true;
					this.infoPanel = null;
					this.active = false;
				} else if (!e.response) {
					this.errormsg = "Impossibile contattare il server.";
				} else if (e.response.status !== 401) {
					this.errormsg = `Errore durante l'aggiornamento (errore ${e.response.status}).`;
				}
			}
		},
		async poll() {
			await this.refresh(false);
			if (this.active) {
				this.timer = setTimeout(this.poll, refreshInterval);
			}
		},
		async afterChange(forceScroll) {
			await this.refresh(forceScroll);
			this.$emit('changed');
		},
		describe(e, action) {
			if (!e.response) {
				return "Impossibile contattare il server.";
			}
			return `${action} non riuscito (errore ${e.response.status}).`;
		},
		async react(message, emoji) {
			try {
				await this.$axios.put(`/conversations/${this.conversationId}/messages/${message.id}/comments/me`, {emoji: emoji});
				await this.afterChange(false);
			} catch (e) {
				this.errormsg = this.describe(e, "Reazione");
			}
		},
		async unreact(message) {
			try {
				await this.$axios.delete(`/conversations/${this.conversationId}/messages/${message.id}/comments/me`);
				await this.afterChange(false);
			} catch (e) {
				this.errormsg = this.describe(e, "Rimozione della reazione");
			}
		},
		async remove(message) {
			if (!window.confirm("Eliminare il messaggio?")) {
				return;
			}
			try {
				await this.$axios.delete(`/conversations/${this.conversationId}/messages/${message.id}`);
				await this.afterChange(false);
			} catch (e) {
				this.errormsg = this.describe(e, "Eliminazione");
			}
		},
		sent() {
			this.replyTo = null;
			this.afterChange(true);
		},
		forwarded(targetId) {
			this.forwarding = null;
			this.$emit('changed');
			this.$router.push(`/conversations/${targetId}`);
		},
		openInfo() {
			if (!this.conversation) {
				return;
			}
			if (this.conversation.type === 'group') {
				this.infoPanel = {type: 'group'};
			} else if (this.otherMemberId !== null) {
				this.infoPanel = {type: 'user', userId: this.otherMemberId};
			}
		},
		openUser(userId) {
			if (userId !== this.myId) {
				this.infoPanel = {type: 'user', userId: userId};
			}
		},
		onKeydown(event) {
			if (event.key === 'Escape') {
				this.viewedPhoto = null;
			}
		},
	},
	mounted() {
		window.addEventListener('keydown', this.onKeydown);
		this.active = true;
		this.poll();
	},
	beforeUnmount() {
		window.removeEventListener('keydown', this.onKeydown);
		this.active = false;
		clearTimeout(this.timer);
	},
}
</script>

<template>
	<div class="chat-view d-flex h-100">
		<div v-if="notFound" class="m-auto text-center p-3">
			<p class="text-body-secondary">Questa conversazione non è disponibile.</p>
			<RouterLink to="/" class="btn btn-outline-primary btn-sm">Torna alla lista</RouterLink>
		</div>

		<div v-else class="chat-main d-flex flex-column flex-grow-1">
			<header class="d-flex align-items-center gap-2 px-2 py-2 border-bottom bg-body">
				<RouterLink to="/" class="btn btn-sm btn-outline-secondary d-md-none" title="Indietro">
					<i class="bi bi-arrow-left"></i>
				</RouterLink>
				<button v-if="conversation" type="button"
					class="chat-header-button d-flex align-items-center gap-2 border-0 bg-transparent p-0 text-start text-body"
					title="Informazioni" @click="openInfo">
					<UserAvatar :picture="conversation.displayPicture" :name="conversation.displayName" :size="40" />
					<div class="chat-title">
						<div class="fw-semibold text-truncate">{{ conversation.displayName }}</div>
						<div v-if="membersText" class="small text-body-secondary text-truncate">{{ membersText }}</div>
					</div>
				</button>
			</header>

			<ErrorMsg v-if="errormsg" :msg="errormsg" class="m-2 py-2"></ErrorMsg>

			<div ref="messageList" class="message-list flex-grow-1 overflow-auto px-3 py-2">
				<LoadingSpinner :loading="conversation === null && !errormsg">
					<MessageItem v-for="m in messages" :key="m.id" :message="m"
						:quoted="m.inReplyTo ? messagesById[m.inReplyTo] || null : null" :my-id="myId"
						@reply="replyTo = $event" @forward="forwarding = $event" @react="react" @unreact="unreact"
						@delete="remove" @open-photo="viewedPhoto = $event" @open-user="openUser" />
				</LoadingSpinner>
			</div>

			<MessageComposer v-if="conversation" :conversation-id="conversationId" :reply-to="replyTo" :my-id="myId"
				@sent="sent" @cancel-reply="replyTo = null" />
		</div>

		<aside v-if="infoPanel && conversation && !notFound" class="info-panel border-start bg-body">
			<GroupInfoPanel v-if="infoPanel.type === 'group'" :conversation="conversation" :my-id="myId"
				@close="infoPanel = null" @changed="afterChange(false)" @open-user="openUser" />
			<UserProfilePanel v-else :key="infoPanel.userId" :user-id="infoPanel.userId" @close="infoPanel = null" />
		</aside>

		<ForwardDialog v-if="forwarding" :conversation-id="conversationId" :message="forwarding"
			@close="forwarding = null" @forwarded="forwarded" />

		<div v-if="viewedPhoto" class="photo-viewer d-flex align-items-center justify-content-center p-3"
			title="Chiudi" @click="viewedPhoto = null">
			<img :src="viewedPhoto" alt="Foto" class="photo-viewer-image">
		</div>
	</div>
</template>

<style>
.chat-view {
	min-height: 0;
}

.chat-main {
	min-width: 0;
}

.chat-header-button {
	min-width: 0;
}

.chat-title {
	min-width: 0;
}

.info-panel {
	width: 340px;
	flex-shrink: 0;
}

@media (max-width: 991.98px) {
	.info-panel {
		position: fixed;
		inset: 0;
		width: auto;
		z-index: 1040;
	}
}

.message-list {
	min-height: 0;
}

.photo-viewer {
	position: fixed;
	inset: 0;
	z-index: 1060;
	background-color: rgba(0, 0, 0, 0.85);
	cursor: zoom-out;
}

.photo-viewer-image {
	max-width: 100%;
	max-height: 100%;
	object-fit: contain;
}
</style>
