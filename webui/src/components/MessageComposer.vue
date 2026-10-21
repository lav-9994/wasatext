<script>
import {checkFile, readAsDataURL, fileToAttachment} from '../services/files.js'

const maxTextLength = 2000;
const uploadTimeout = 60000;

export default {
	props: {
		conversationId: {type: Number, required: true},
		replyTo: {type: Object, default: null},
		myId: {type: Number, required: true},
	},
	emits: ['sent', 'cancel-reply'],
	data() {
		return {
			text: '',
			photo: null,
			file: null,
			sending: false,
			errormsg: null,
		}
	},
	computed: {
		textLength() {
			return [...this.text.trim()].length;
		},
		tooLong() {
			return this.textLength > maxTextLength;
		},
		canSend() {
			return !this.sending && !this.tooLong && (this.photo !== null || this.file !== null || this.textLength > 0);
		},
		replySender() {
			if (!this.replyTo) {
				return "";
			}
			return this.replyTo.sender.id === this.myId ? "Tu" : this.replyTo.sender.username;
		},
		replyPreview() {
			if (!this.replyTo) {
				return "";
			}
			if (this.replyTo.text) {
				return this.replyTo.text;
			}
			if (this.replyTo.type === 'photo') {
				return "Foto";
			}
			if (this.replyTo.type === 'file') {
				return this.replyTo.file.name;
			}
			return "";
		},
	},
	watch: {
		replyTo(value) {
			if (value) {
				this.$nextTick(() => this.$refs.textInput.focus());
			}
		},
	},
	methods: {
		choosePhoto() {
			this.$refs.photoInput.click();
		},
		chooseFile() {
			this.$refs.fileInput.click();
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
			try {
				this.photo = await readAsDataURL(chosen);
				this.file = null;
			} catch (e) {
				this.errormsg = "Impossibile leggere l'immagine.";
			}
		},
		async fileChosen(event) {
			const chosen = event.target.files[0];
			event.target.value = '';
			if (!chosen) {
				return;
			}
			this.errormsg = checkFile(chosen, false);
			if (this.errormsg) {
				return;
			}
			try {
				this.file = await fileToAttachment(chosen);
				this.photo = null;
			} catch (e) {
				this.errormsg = "Impossibile leggere il file.";
			}
		},
		removeAttachment() {
			this.photo = null;
			this.file = null;
		},
		resize() {
			const input = this.$refs.textInput;
			input.style.height = 'auto';
			input.style.height = `${input.scrollHeight}px`;
		},
		async send() {
			if (!this.canSend) {
				return;
			}
			const body = {type: 'text'};
			const text = this.text.trim();
			if (text !== '') {
				body.text = text;
			}
			if (this.photo !== null) {
				body.type = 'photo';
				body.photo = this.photo;
			} else if (this.file !== null) {
				body.type = 'file';
				body.file = this.file;
			}
			if (this.replyTo) {
				body.inReplyTo = this.replyTo.id;
			}
			this.sending = true;
			this.errormsg = null;
			try {
				await this.$axios.post(`/conversations/${this.conversationId}/messages`, body, {timeout: uploadTimeout});
				this.text = '';
				this.photo = null;
				this.file = null;
				this.$nextTick(this.resize);
				this.$emit('sent');
			} catch (e) {
				if (!e.response) {
					this.errormsg = "Impossibile contattare il server.";
				} else if (e.response.status === 400) {
					this.errormsg = "Il messaggio non è valido.";
				} else {
					this.errormsg = `Invio non riuscito (errore ${e.response.status}).`;
				}
			}
			this.sending = false;
		},
	},
}
</script>

<template>
	<div class="composer border-top bg-body p-2">
		<ErrorMsg v-if="errormsg" :msg="errormsg" class="mb-2 py-2"></ErrorMsg>

		<div v-if="replyTo" class="composer-extra d-flex align-items-center gap-2 mb-2 p-2 rounded">
			<i class="bi bi-reply text-body-secondary"></i>
			<div class="flex-grow-1 composer-extra-body small">
				<div class="fw-semibold">Rispondi a {{ replySender }}</div>
				<div class="text-truncate text-body-secondary">{{ replyPreview }}</div>
			</div>
			<button type="button" class="btn btn-sm" title="Annulla risposta" @click="$emit('cancel-reply')">
				<i class="bi bi-x-lg"></i>
			</button>
		</div>

		<div v-if="photo || file" class="composer-extra d-flex align-items-center gap-2 mb-2 p-2 rounded">
			<img v-if="photo" :src="photo" alt="Anteprima" class="composer-photo rounded">
			<template v-else>
				<i class="bi bi-file-earmark fs-3"></i>
			</template>
			<div class="flex-grow-1 composer-extra-body small">
				<div class="fw-semibold">{{ photo ? 'Foto' : 'File' }}</div>
				<div v-if="file" class="text-truncate text-body-secondary">{{ file.name }}</div>
				<div class="text-body-secondary">Il testo che scrivi sarà la didascalia</div>
			</div>
			<button type="button" class="btn btn-sm" title="Rimuovi allegato" @click="removeAttachment">
				<i class="bi bi-x-lg"></i>
			</button>
		</div>

		<div class="d-flex align-items-end gap-1">
			<button type="button" class="btn btn-outline-secondary" title="Invia una foto" :disabled="sending"
				@click="choosePhoto">
				<i class="bi bi-image"></i>
			</button>
			<button type="button" class="btn btn-outline-secondary" title="Invia un file" :disabled="sending"
				@click="chooseFile">
				<i class="bi bi-paperclip"></i>
			</button>
			<input ref="photoInput" type="file" accept="image/*" class="d-none" @change="photoChosen">
			<input ref="fileInput" type="file" class="d-none" @change="fileChosen">

			<textarea ref="textInput" v-model="text" rows="1" class="form-control composer-input"
				:class="{'is-invalid': tooLong}" placeholder="Scrivi un messaggio" @input="resize"
				@keydown.enter.exact.prevent="send"></textarea>

			<button type="button" class="btn btn-primary" title="Invia" :disabled="!canSend" @click="send">
				<span v-if="sending" class="spinner-border spinner-border-sm"></span>
				<i v-else class="bi bi-send"></i>
			</button>
		</div>
		<div v-if="tooLong" class="small text-danger mt-1">Il messaggio supera i 2000 caratteri.</div>
	</div>
</template>

<style>
.composer-input {
	resize: none;
	max-height: 140px;
	overflow-y: auto;
}

.composer-extra {
	background-color: var(--bs-tertiary-bg);
}

.composer-extra-body {
	min-width: 0;
}

.composer-photo {
	width: 56px;
	height: 56px;
	object-fit: cover;
}
</style>
