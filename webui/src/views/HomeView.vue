<script>
import {getUserId, clearUserId} from '../services/session.js'
import ChatList from '../components/ChatList.vue'
import UserAvatar from '../components/UserAvatar.vue'
import ChatView from '../components/ChatView.vue'
import MyProfilePanel from '../components/MyProfilePanel.vue'
import NewChatPanel from '../components/NewChatPanel.vue'
import NewGroupPanel from '../components/NewGroupPanel.vue'

const refreshInterval = 2500;

export default {
	components: {ChatList, UserAvatar, ChatView, MyProfilePanel, NewChatPanel, NewGroupPanel},
	data() {
		return {
			myId: getUserId(),
			me: null,
			conversations: [],
			loaded: false,
			leftPanel: 'chats',
			errormsg: null,
			active: false,
			timer: null,
		}
	},
	computed: {
		selectedId() {
			const id = Number(this.$route.params.id);
			return Number.isInteger(id) && id > 0 ? id : null;
		},
	},
	watch: {
		selectedId() {
			this.refresh();
		},
	},
	methods: {
		async refresh() {
			try {
				const [profile, list] = await Promise.all([
					this.$axios.get(`/users/${this.myId}`),
					this.$axios.get("/conversations"),
				]);
				this.me = profile.data;
				this.conversations = list.data;
				this.errormsg = null;
			} catch (e) {
				if (!e.response) {
					this.errormsg = "Impossibile contattare il server.";
				} else if (e.response.status !== 401) {
					this.errormsg = `Errore durante l'aggiornamento (errore ${e.response.status}).`;
				}
			}
			this.loaded = true;
		},
		async poll() {
			await this.refresh();
			if (this.active) {
				this.timer = setTimeout(this.poll, refreshInterval);
			}
		},
		panelDone() {
			this.leftPanel = 'chats';
			this.refresh();
		},
		logout() {
			clearUserId();
			this.$router.replace("/login");
		},
	},
	mounted() {
		this.active = true;
		this.poll();
	},
	beforeUnmount() {
		this.active = false;
		clearTimeout(this.timer);
	},
}
</script>

<template>
	<div class="home d-flex" :class="{'chat-open': selectedId !== null}">
		<aside class="chats-column d-flex flex-column border-end bg-body">
			<MyProfilePanel v-if="leftPanel === 'profile'" :me="me" @back="leftPanel = 'chats'" @changed="refresh"
				@logout="logout" />
			<NewChatPanel v-else-if="leftPanel === 'newChat'" @back="leftPanel = 'chats'" @opened="panelDone" />
			<NewGroupPanel v-else-if="leftPanel === 'newGroup'" @back="leftPanel = 'chats'" @created="panelDone" />

			<header v-if="leftPanel === 'chats'" class="d-flex align-items-center gap-2 p-2 border-bottom">
				<button type="button"
					class="profile-summary d-flex align-items-center gap-2 border-0 bg-transparent p-0 text-start text-body"
					title="Il tuo profilo" @click="leftPanel = 'profile'">
					<UserAvatar :picture="me ? me.profilePicture : null" :name="me ? me.username : ''" :size="40" />
					<span class="fw-semibold text-truncate">{{ me ? me.username : '' }}</span>
				</button>
				<div class="ms-auto d-flex gap-1">
					<button class="btn btn-sm btn-outline-secondary" type="button" title="Nuova chat"
						@click="leftPanel = 'newChat'">
						<i class="bi bi-chat-text"></i>
					</button>
					<button class="btn btn-sm btn-outline-secondary" type="button" title="Nuovo gruppo"
						@click="leftPanel = 'newGroup'">
						<i class="bi bi-people"></i>
					</button>
					<button class="btn btn-sm btn-outline-secondary" type="button"
						:title="$root.theme === 'dark' ? 'Tema chiaro' : 'Tema scuro'" @click="$root.toggleTheme()">
						<i :class="$root.theme === 'dark' ? 'bi bi-sun' : 'bi bi-moon-stars'"></i>
					</button>
					<button class="btn btn-sm btn-outline-secondary" type="button" title="Esci" @click="logout">
						<i class="bi bi-box-arrow-right"></i>
					</button>
				</div>
			</header>

			<ErrorMsg v-if="errormsg && leftPanel === 'chats'" :msg="errormsg" class="m-2"></ErrorMsg>

			<ChatList v-if="leftPanel === 'chats'" class="chats-list flex-grow-1" :conversations="conversations"
				:selected-id="selectedId" :my-id="myId" :loaded="loaded" />
		</aside>

		<main class="conversation-column flex-grow-1 d-flex flex-column bg-body-tertiary">
			<div v-if="selectedId === null" class="m-auto text-body-secondary">
				Seleziona una chat per iniziare
			</div>
			<ChatView v-else :key="selectedId" :conversation-id="selectedId" :my-id="myId" @changed="refresh" />
		</main>
	</div>
</template>

<style>
.home {
	height: 100vh;
	height: 100dvh;
}

.chats-column {
	width: 380px;
	flex-shrink: 0;
}

.chats-list {
	min-height: 0;
}

.profile-summary {
	min-width: 0;
}

@media (max-width: 767.98px) {
	.chats-column {
		width: 100%;
	}

	.home.chat-open .chats-column {
		display: none !important;
	}

	.home:not(.chat-open) .conversation-column {
		display: none !important;
	}
}
</style>
