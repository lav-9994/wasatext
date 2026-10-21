<script>
import UserAvatar from './UserAvatar.vue'
import UserSearchList from './UserSearchList.vue'
import {checkFile, readAsDataURL} from '../services/files.js'

export default {
	components: {UserAvatar, UserSearchList},
	props: {
		conversation: {type: Object, required: true},
		myId: {type: Number, required: true},
	},
	emits: ['close', 'changed', 'open-user'],
	data() {
		return {
			editingName: false,
			nameDraft: '',
			editingDescription: false,
			descriptionDraft: '',
			adding: false,
			toAdd: [],
			saving: false,
			errormsg: null,
		}
	},
	computed: {
		groupUrl() {
			return `/groups/${this.conversation.id}`;
		},
		nameLength() {
			return [...this.nameDraft.trim()].length;
		},
		nameValid() {
			return this.nameLength >= 1 && this.nameLength <= 32;
		},
		descriptionLength() {
			return [...this.descriptionDraft].length;
		},
		memberIds() {
			return this.conversation.members.map((m) => m.id);
		},
	},
	methods: {
		describe(e, action) {
			if (!e.response) {
				return "Impossibile contattare il server.";
			}
			return `${action} non riuscito (errore ${e.response.status}).`;
		},
		startName() {
			this.nameDraft = this.conversation.displayName;
			this.editingName = true;
		},
		async saveName() {
			if (!this.nameValid || this.saving) {
				return;
			}
			this.saving = true;
			this.errormsg = null;
			try {
				await this.$axios.put(`${this.groupUrl}/name`, {name: this.nameDraft.trim()});
				this.editingName = false;
				this.$emit('changed');
			} catch (e) {
				this.errormsg = this.describe(e, "Salvataggio del nome");
			}
			this.saving = false;
		},
		startDescription() {
			this.descriptionDraft = this.conversation.description || '';
			this.editingDescription = true;
		},
		async saveDescription() {
			if (this.descriptionLength > 150 || this.saving) {
				return;
			}
			this.saving = true;
			this.errormsg = null;
			try {
				await this.$axios.put(`${this.groupUrl}/description`, {description: this.descriptionDraft.trim()});
				this.editingDescription = false;
				this.$emit('changed');
			} catch (e) {
				this.errormsg = this.describe(e, "Salvataggio della descrizione");
			}
			this.saving = false;
		},
		choosePhoto() {
			this.$refs.photoInput.click();
		},
		async photoChosen(event) {
			const chosen = event.target.files[0];
			event.target.value = '';
			if (!chosen) {
				return;
			}
			this.errormsg = checkFile(chosen, true);
			if (this.errormsg) {
				return;
			}
			this.saving = true;
			try {
				const picture = await readAsDataURL(chosen);
				await this.$axios.put(`${this.groupUrl}/photo`, {picture: picture}, {timeout: 60000});
				this.$emit('changed');
			} catch (e) {
				this.errormsg = this.describe(e, "Caricamento della foto");
			}
			this.saving = false;
		},
		startAdding() {
			this.toAdd = [];
			this.adding = true;
		},
		toggleAdd(user) {
			const index = this.toAdd.findIndex((u) => u.id === user.id);
			if (index >= 0) {
				this.toAdd.splice(index, 1);
			} else {
				this.toAdd.push(user);
			}
		},
		async addMembers() {
			if (this.toAdd.length === 0 || this.saving) {
				return;
			}
			this.saving = true;
			this.errormsg = null;
			try {
				await this.$axios.post(`${this.groupUrl}/members`, {userIds: this.toAdd.map((u) => u.id)});
				this.adding = false;
				this.$emit('changed');
			} catch (e) {
				this.errormsg = this.describe(e, "Aggiunta dei membri");
			}
			this.saving = false;
		},
		async leave() {
			if (!window.confirm("Vuoi davvero uscire dal gruppo?")) {
				return;
			}
			this.saving = true;
			this.errormsg = null;
			try {
				await this.$axios.delete(`${this.groupUrl}/members/me`);
				this.$emit('changed');
				this.$router.push("/");
			} catch (e) {
				this.errormsg = this.describe(e, "Uscita dal gruppo");
			}
			this.saving = false;
		},
	},
}
</script>

<template>
	<div class="d-flex flex-column h-100">
		<header class="d-flex align-items-center gap-2 p-2 border-bottom">
			<span class="fw-semibold">Info gruppo</span>
			<button type="button" class="btn-close ms-auto" title="Chiudi" @click="$emit('close')"></button>
		</header>

		<div class="flex-grow-1 overflow-auto p-3">
			<ErrorMsg v-if="errormsg" :msg="errormsg" class="py-2"></ErrorMsg>

			<div class="d-flex flex-column align-items-center mb-3">
				<UserAvatar :picture="conversation.displayPicture" :name="conversation.displayName" :size="112" />
				<button type="button" class="btn btn-sm btn-outline-primary mt-2" :disabled="saving" @click="choosePhoto">
					<i class="bi bi-camera me-1"></i> Cambia foto
				</button>
				<input ref="photoInput" type="file" accept="image/*" class="d-none" @change="photoChosen">
			</div>

			<div class="mb-3">
				<div class="small text-body-secondary mb-1">Nome</div>
				<form v-if="editingName" novalidate @submit.prevent="saveName">
					<input v-model="nameDraft" type="text" class="form-control"
						:class="{'is-invalid': nameDraft !== '' && !nameValid}" autofocus>
					<div class="invalid-feedback">Il nome deve avere tra 1 e 32 caratteri.</div>
					<div class="d-flex gap-2 mt-2">
						<button type="submit" class="btn btn-sm btn-primary" :disabled="!nameValid || saving">Salva</button>
						<button type="button" class="btn btn-sm btn-outline-secondary" @click="editingName = false">Annulla</button>
					</div>
				</form>
				<div v-else class="d-flex align-items-center gap-2">
					<span class="fs-5 fw-semibold text-break">{{ conversation.displayName }}</span>
					<button type="button" class="btn btn-sm btn-link ms-auto" title="Modifica nome" @click="startName">
						<i class="bi bi-pencil"></i>
					</button>
				</div>
			</div>

			<div class="mb-3">
				<div class="small text-body-secondary mb-1">Descrizione</div>
				<form v-if="editingDescription" novalidate @submit.prevent="saveDescription">
					<textarea v-model="descriptionDraft" rows="3" class="form-control"
						:class="{'is-invalid': descriptionLength > 150}" placeholder="Descrizione del gruppo"></textarea>
					<div class="form-text">{{ descriptionLength }}/150 · lascia vuoto per rimuoverla</div>
					<div class="d-flex gap-2 mt-2">
						<button type="submit" class="btn btn-sm btn-primary"
							:disabled="descriptionLength > 150 || saving">Salva</button>
						<button type="button" class="btn btn-sm btn-outline-secondary"
							@click="editingDescription = false">Annulla</button>
					</div>
				</form>
				<div v-else class="d-flex align-items-start gap-2">
					<span v-if="conversation.description" class="text-break bio-text">{{ conversation.description }}</span>
					<span v-else class="text-body-secondary fst-italic">Nessuna descrizione</span>
					<button type="button" class="btn btn-sm btn-link ms-auto" title="Modifica descrizione"
						@click="startDescription">
						<i class="bi bi-pencil"></i>
					</button>
				</div>
			</div>

			<p v-if="conversation.creator" class="small text-body-secondary">
				Creato da {{ conversation.creator.id === myId ? 'te' : conversation.creator.username }}
			</p>

			<div class="d-flex align-items-center mb-1">
				<span class="small text-body-secondary">{{ conversation.members.length }} membri</span>
				<button type="button" class="btn btn-sm btn-outline-primary ms-auto" :disabled="saving" @click="startAdding">
					<i class="bi bi-person-plus me-1"></i> Aggiungi membri
				</button>
			</div>
			<button v-for="m in conversation.members" :key="m.id" type="button"
				class="member-item d-flex align-items-center gap-2 w-100 border-0 bg-transparent py-1 px-0 text-start text-body"
				:disabled="m.id === myId" @click="$emit('open-user', m.id)">
				<UserAvatar :picture="m.profilePicture" :name="m.username" :size="36" />
				<span class="text-truncate">{{ m.id === myId ? m.username + ' (tu)' : m.username }}</span>
			</button>

			<button type="button" class="btn btn-outline-danger w-100 mt-4" :disabled="saving" @click="leave">
				<i class="bi bi-box-arrow-left me-1"></i> Esci dal gruppo
			</button>
		</div>

		<div v-if="adding" class="dialog-backdrop d-flex align-items-center justify-content-center p-3"
			@click.self="adding = false">
			<div class="dialog-card card shadow">
				<div class="card-header d-flex align-items-center gap-2">
					<span class="fw-semibold">Aggiungi membri</span>
					<button type="button" class="btn-close ms-auto" title="Chiudi" @click="adding = false"></button>
				</div>
				<UserSearchList class="dialog-body flex-grow-1" selectable :selected-ids="toAdd.map((u) => u.id)"
					:exclude-ids="memberIds" @pick="toggleAdd" />
				<div class="card-footer">
					<button type="button" class="btn btn-primary w-100" :disabled="toAdd.length === 0 || saving"
						@click="addMembers">
						Aggiungi{{ toAdd.length > 0 ? ' (' + toAdd.length + ')' : '' }}
					</button>
				</div>
			</div>
		</div>
	</div>
</template>

<style>
.member-item:disabled {
	opacity: 1;
}
</style>
