<script>
import UserAvatar from './UserAvatar.vue'
import {checkFile, readAsDataURL} from '../services/files.js'

export default {
	components: {UserAvatar},
	props: {
		me: {type: Object, default: null},
	},
	emits: ['back', 'changed', 'logout'],
	data() {
		return {
			editingName: false,
			nameDraft: '',
			nameError: null,
			editingBio: false,
			bioDraft: '',
			saving: false,
			errormsg: null,
		}
	},
	computed: {
		nameLength() {
			return [...this.nameDraft.trim()].length;
		},
		nameValid() {
			return this.nameLength >= 3 && this.nameLength <= 16;
		},
		bioLength() {
			return [...this.bioDraft].length;
		},
	},
	methods: {
		startName() {
			this.nameDraft = this.me.username;
			this.nameError = null;
			this.editingName = true;
		},
		async saveName() {
			if (!this.nameValid || this.saving) {
				return;
			}
			this.saving = true;
			this.nameError = null;
			try {
				await this.$axios.put("/users/me/username", {username: this.nameDraft.trim()});
				this.editingName = false;
				this.$emit('changed');
			} catch (e) {
				if (e.response && e.response.status === 409) {
					this.nameError = "Questo nome è già usato da un altro utente.";
				} else if (e.response && e.response.status === 400) {
					this.nameError = "Il nome deve avere tra 3 e 16 caratteri.";
				} else {
					this.nameError = e.response ? `Salvataggio non riuscito (errore ${e.response.status}).` : "Impossibile contattare il server.";
				}
			}
			this.saving = false;
		},
		startBio() {
			this.bioDraft = this.me.bio || '';
			this.editingBio = true;
		},
		async saveBio() {
			if (this.bioLength > 150 || this.saving) {
				return;
			}
			this.saving = true;
			this.errormsg = null;
			try {
				await this.$axios.put("/users/me/bio", {bio: this.bioDraft.trim()});
				this.editingBio = false;
				this.$emit('changed');
			} catch (e) {
				this.errormsg = e.response ? `Salvataggio non riuscito (errore ${e.response.status}).` : "Impossibile contattare il server.";
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
				await this.$axios.put("/users/me/photo", {profilePicture: picture}, {timeout: 60000});
				this.$emit('changed');
			} catch (e) {
				this.errormsg = e.response ? `Caricamento non riuscito (errore ${e.response.status}).` : "Impossibile caricare la foto.";
			}
			this.saving = false;
		},
	},
}
</script>

<template>
	<div class="d-flex flex-column h-100">
		<header class="d-flex align-items-center gap-2 p-2 border-bottom">
			<button type="button" class="btn btn-sm btn-outline-secondary" title="Indietro" @click="$emit('back')">
				<i class="bi bi-arrow-left"></i>
			</button>
			<span class="fw-semibold">Il tuo profilo</span>
		</header>

		<div v-if="me" class="flex-grow-1 overflow-auto p-3">
			<ErrorMsg v-if="errormsg" :msg="errormsg" class="py-2"></ErrorMsg>

			<div class="d-flex flex-column align-items-center mb-4">
				<UserAvatar :picture="me.profilePicture" :name="me.username" :size="112" />
				<button type="button" class="btn btn-sm btn-outline-primary mt-2" :disabled="saving" @click="choosePhoto">
					<i class="bi bi-camera me-1"></i> Cambia foto
				</button>
				<input ref="photoInput" type="file" accept="image/*" class="d-none" @change="photoChosen">
			</div>

			<div class="mb-4">
				<div class="small text-body-secondary mb-1">Username</div>
				<form v-if="editingName" novalidate @submit.prevent="saveName">
					<input v-model="nameDraft" type="text" class="form-control"
						:class="{'is-invalid': nameError || (nameDraft.trim() !== '' && !nameValid)}" autofocus>
					<div class="invalid-feedback">{{ nameError || "Il nome deve avere tra 3 e 16 caratteri." }}</div>
					<div class="d-flex gap-2 mt-2">
						<button type="submit" class="btn btn-sm btn-primary" :disabled="!nameValid || saving">Salva</button>
						<button type="button" class="btn btn-sm btn-outline-secondary" @click="editingName = false">Annulla</button>
					</div>
				</form>
				<div v-else class="d-flex align-items-center gap-2">
					<span class="fs-5 fw-semibold text-break">{{ me.username }}</span>
					<button type="button" class="btn btn-sm btn-link ms-auto" title="Modifica username" @click="startName">
						<i class="bi bi-pencil"></i>
					</button>
				</div>
			</div>

			<div class="mb-4">
				<div class="small text-body-secondary mb-1">Bio</div>
				<form v-if="editingBio" novalidate @submit.prevent="saveBio">
					<textarea v-model="bioDraft" rows="3" class="form-control" :class="{'is-invalid': bioLength > 150}"
						placeholder="Scrivi qualcosa su di te"></textarea>
					<div class="form-text">{{ bioLength }}/150</div>
					<div class="d-flex gap-2 mt-2">
						<button type="submit" class="btn btn-sm btn-primary" :disabled="bioLength > 150 || saving">Salva</button>
						<button type="button" class="btn btn-sm btn-outline-secondary" @click="editingBio = false">Annulla</button>
					</div>
				</form>
				<div v-else class="d-flex align-items-start gap-2">
					<span v-if="me.bio" class="text-break bio-text">{{ me.bio }}</span>
					<span v-else class="text-body-secondary fst-italic">Nessuna bio</span>
					<button type="button" class="btn btn-sm btn-link ms-auto" title="Modifica bio" @click="startBio">
						<i class="bi bi-pencil"></i>
					</button>
				</div>
			</div>

			<button type="button" class="btn btn-outline-danger w-100" @click="$emit('logout')">
				<i class="bi bi-box-arrow-right me-1"></i> Esci
			</button>
		</div>
	</div>
</template>

<style>
.bio-text {
	white-space: pre-wrap;
}
</style>
