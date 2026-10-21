<script>
import {saveUserId} from '../services/session.js'

export default {
	data() {
		return {
			username: '',
			errormsg: null,
			loading: false,
		}
	},
	computed: {
		usernameLength() {
			return [...this.username.trim()].length;
		},
		validationError() {
			if (this.usernameLength > 0 && (this.usernameLength < 3 || this.usernameLength > 16)) {
				return "Lo username deve avere tra 3 e 16 caratteri.";
			}
			return null;
		},
		canSubmit() {
			return !this.loading && this.usernameLength >= 3 && this.usernameLength <= 16;
		},
	},
	methods: {
		async login() {
			if (!this.canSubmit) {
				return;
			}
			this.loading = true;
			this.errormsg = null;
			try {
				const response = await this.$axios.post("/session", {name: this.username.trim()});
				saveUserId(response.data.identifier);
				this.$router.replace("/");
			} catch (e) {
				if (e.response) {
					this.errormsg = `Accesso non riuscito (errore ${e.response.status}).`;
				} else {
					this.errormsg = "Impossibile contattare il server.";
				}
			}
			this.loading = false;
		},
	},
}
</script>

<template>
	<div class="min-vh-100 d-flex align-items-center justify-content-center bg-body-tertiary p-3">
		<div class="card shadow-sm w-100 login-card">
			<div class="card-body p-4">
				<h1 class="h3 text-center mb-1">
					<i class="bi bi-chat-dots-fill text-primary"></i> WASAText
				</h1>
				<p class="text-center text-body-secondary mb-4">Inserisci il tuo username per entrare</p>

				<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

				<form novalidate @submit.prevent="login">
					<label for="usernameInput" class="form-label">Username</label>
					<input id="usernameInput" v-model="username" type="text" class="form-control"
						:class="{'is-invalid': validationError}" autocomplete="username" autofocus>
					<div class="invalid-feedback">{{ validationError }}</div>
					<button type="submit" class="btn btn-primary w-100 mt-3" :disabled="!canSubmit">
						Entra
					</button>
				</form>
			</div>
		</div>
	</div>
</template>

<style>
.login-card {
	max-width: 380px;
}
</style>
