<script>
import UserSearchList from './UserSearchList.vue'

const maxMembers = 499;

export default {
	components: {UserSearchList},
	emits: ['back', 'created'],
	data() {
		return {
			step: 'members',
			selected: [],
			name: '',
			creating: false,
			errormsg: null,
		}
	},
	computed: {
		selectedIds() {
			return this.selected.map((u) => u.id);
		},
		nameLength() {
			return [...this.name.trim()].length;
		},
		canCreate() {
			return !this.creating && this.selected.length > 0 && this.nameLength >= 1 && this.nameLength <= 32;
		},
	},
	methods: {
		toggle(user) {
			const index = this.selected.findIndex((u) => u.id === user.id);
			if (index >= 0) {
				this.selected.splice(index, 1);
			} else if (this.selected.length < maxMembers) {
				this.selected.push(user);
			}
		},
		back() {
			if (this.step === 'name') {
				this.step = 'members';
			} else {
				this.$emit('back');
			}
		},
		async create() {
			if (!this.canCreate) {
				return;
			}
			this.creating = true;
			this.errormsg = null;
			try {
				const response = await this.$axios.post("/groups", {name: this.name.trim(), members: this.selectedIds});
				this.$emit('created');
				this.$router.push(`/conversations/${response.data.id}`);
			} catch (e) {
				this.errormsg = e.response ? `Creazione non riuscita (errore ${e.response.status}).` : "Impossibile contattare il server.";
			}
			this.creating = false;
		},
	},
}
</script>

<template>
	<div class="d-flex flex-column h-100">
		<header class="d-flex align-items-center gap-2 p-2 border-bottom">
			<button type="button" class="btn btn-sm btn-outline-secondary" title="Indietro" @click="back">
				<i class="bi bi-arrow-left"></i>
			</button>
			<span class="fw-semibold">Nuovo gruppo</span>
			<button v-if="step === 'members'" type="button" class="btn btn-sm btn-primary ms-auto"
				:disabled="selected.length === 0" @click="step = 'name'">
				Avanti
			</button>
		</header>

		<template v-if="step === 'members'">
			<div v-if="selected.length > 0" class="d-flex flex-wrap gap-1 p-2 border-bottom">
				<span v-for="u in selected" :key="u.id"
					class="badge rounded-pill bg-body-secondary text-body border d-flex align-items-center gap-1">
					{{ u.username }}
					<button type="button" class="btn-close btn-close-sm" :title="'Togli ' + u.username"
						@click="toggle(u)"></button>
				</span>
			</div>
			<p v-else class="small text-body-secondary px-3 pt-2 mb-0">Scegli i membri del gruppo</p>
			<UserSearchList class="panel-fill flex-grow-1" selectable :selected-ids="selectedIds" @pick="toggle" />
		</template>

		<form v-else class="p-3" novalidate @submit.prevent="create">
			<ErrorMsg v-if="errormsg" :msg="errormsg" class="py-2"></ErrorMsg>
			<label for="groupNameInput" class="form-label">Nome del gruppo</label>
			<input id="groupNameInput" v-model="name" type="text" class="form-control"
				:class="{'is-invalid': nameLength > 32}" autofocus>
			<div class="invalid-feedback">Il nome può avere al massimo 32 caratteri.</div>
			<p class="small text-body-secondary mt-2">
				Membri: {{ selected.map((u) => u.username).join(', ') }}
			</p>
			<button type="submit" class="btn btn-primary w-100" :disabled="!canCreate">
				<span v-if="creating" class="spinner-border spinner-border-sm me-1"></span>
				Crea gruppo
			</button>
		</form>
	</div>
</template>

<style>
.btn-close-sm {
	width: 0.5em;
	height: 0.5em;
	padding: 0.25em;
}
</style>
