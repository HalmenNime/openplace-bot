<script setup>
    import { sumBy } from 'lodash';
    import { onMounted, ref, useTemplateRef } from 'vue';
    import { SelectImage, ReadFile, Request, WriteSettings, ReadSettings } from '../wailsjs/go/main/App';
    import { alert, generateId, imageDataFromBuffer, input, isUnsignedInteger, numberFormat, randomstring } from './helpers';

    const BASE_URL = 'https://place34.com/';

    const canvas = useTemplateRef('canvas');
    const settings = ref({
        tileX: null,
        tileY: null,
        pX: null,
        pY: null,
        image: null,
        dithering: false,
        usePremiumColors: false,
        buyCharges: false,
        buyMaxCharges: 0,
        sleep: 60,
        requestConcurrent: 5,
        users: [],
    });
    const logs = ref([]);
    const loading = ref(false);
    const running = ref(false);
    const stopping = ref(false);

    function log(message) {
        const text = `[${new Date().toLocaleTimeString()}] ${message}`;
        logs.value.push(text);
        logs.value.splice(0, logs.value.length - 100);
    }

    function url(path) {
        return BASE_URL.replace(/\/$/, '') + '/' + path.replace(/^\//, '');
    }

    async function selectImage() {
        const filename = await SelectImage();
        if (!filename) return;

        settings.value.image = filename;

        await loadImage();
        await writeSettings();
    }

    async function loadImage() {
        const buffer = await ReadFile(settings.value.image);
        const bytes = Uint8Array.from(atob(buffer), (c) => c.charCodeAt(0));
        const imageData = await imageDataFromBuffer(bytes);

        canvas.value.width = imageData.width;
        canvas.value.height = imageData.height;

        const ctx = canvas.value.getContext('2d');
        ctx.putImageData(imageData, 0, 0);
    }

    async function addUser() {
        let amount = await input('Enter the amount of users to add');
        if (!amount) return;
        if (!isUnsignedInteger(amount)) {
            alert('Invalid amount');
            return;
        }

        loading.value = true;

        log(`Adding ${amount} users...`);
        amount = parseInt(amount);

        const promises = [];
        for (let i = 0; i < settings.value.requestConcurrent; i++) {
            const promise = new Promise(async (resolve, reject) => {
                while (amount > 0) {
                    amount -= 1;

                    const username = `bot-` + randomstring(12);
                    const password = randomstring();

                    try {
                        const user = {
                            id: generateId(),
                            username: username,
                            password: password,
                            enabled: true,
                        };
                        await Login(user);
                        await FetchMe(user);
                        settings.value.users.push(user);
                        log(`User ${username} added.`);
                    } catch (error) {
                        reject(error);
                        return;
                    }
                }

                resolve();
            });
            promises.push(promise);
        }

        try {
            await Promise.all(promises);
        } catch (error) {
            log(error.message);
        }

        await writeSettings();

        loading.value = false;
    }

    async function Login(user) {
        const response = await Request({
            method: 'POST',
            url: url('/login'),
            data: JSON.stringify({ username: user.username, password: user.password }),
        });
        const raw = atob(response.data);

        if (response.status !== 200) {
            throw new Error(`Failed to login with status ${response.status}. Response: ${raw}`);
        }

        user.cookie = (response.headers['Set-Cookie'] || []).map((x) => x.replace(/^(\w+=.*?)(;.*?)*$/, '$1')).join(';');
    }

    async function FetchMe(user) {
        const response = await Request({ method: 'GET', url: url('/me'), cookie: user.cookie });
        const raw = atob(response.data);

        if (response.status !== 200) {
            throw new Error(`Failed to fetch me with status ${response.status}. Response: ${raw}`);
        }

        const data = JSON.parse(raw);
        user.me = {
            banned: data.banned,
            charges: data.charges,
            droplets: data.droplets,
            extraColorsBitmap: data.extraColorsBitmap,
            flagsBitmap: data.flagsBitmap,
            id: data.id,
            name: data.name,
        };
        user.lastFetch = Date.now();
    }

    async function writeSettings() {
        try {
            await WriteSettings(JSON.stringify(settings.value));
            log('Settings saved.');
        } catch (error) {
            log('Failed to save settings: ' + error.message);
        }
    }

    async function readSettings() {
        try {
            const data = await ReadSettings();
            if (!data) return;

            settings.value = JSON.parse(data);
        } catch (error) {
            log('Failed to read settings: ' + error.message);
        }
    }

    async function start() {}

    onMounted(async () => {
        loading.value = true;

        try {
            await readSettings();
            await loadImage();
        } catch (error) {
            log('Failed to read settings: ' + error.message);
        }

        loading.value = false;
    });
</script>

<template>
    <div class="grid-container">
        <div class="grid-item">
            <div class="border rounded p-2 h-100">
                <div class="row g-2 align-items-center">
                    <div class="col-2 text-end">
                        <label for="tileX">Tile X</label>
                    </div>
                    <div class="col-4">
                        <input type="text" v-model="settings.tileX" id="tileX" class="form-control" placeholder="Tile X" :disabled="loading || running" @change="writeSettings" />
                    </div>
                    <div class="col-2 text-end">
                        <label for="tileY">Tile Y</label>
                    </div>
                    <div class="col-4">
                        <input type="text" v-model="settings.tileY" id="tileY" class="form-control" placeholder="Tile Y" :disabled="loading || running" @change="writeSettings" />
                    </div>
                    <div class="col-2 text-end">
                        <label for="pX">P X</label>
                    </div>
                    <div class="col-4">
                        <input type="text" v-model="settings.pX" id="pX" class="form-control" placeholder="P X" :disabled="loading || running" @change="writeSettings" />
                    </div>
                    <div class="col-2 text-end">
                        <label for="pY">P Y</label>
                    </div>
                    <div class="col-4">
                        <input type="text" v-model="settings.pY" id="pY" class="form-control" placeholder="P Y" :disabled="loading || running" @change="writeSettings" />
                    </div>
                    <div class="col-2 text-end align-self-start">
                        <label for="tileX">Image</label>
                    </div>
                    <div class="col-10">
                        <div class="border rounded d-flex flex-column">
                            <canvas ref="canvas"></canvas>
                            <button type="button" class="btn btn-sm btn-primary border-0 rounded-top-0" @click="selectImage" :disabled="loading">
                                <i class="fa-solid fa-image"></i>
                                Select image
                            </button>
                        </div>
                    </div>
                    <div class="col-2 text-end align-self-start">
                        <label for="tileX">Options</label>
                    </div>
                    <div class="col-10">
                        <div class="d-flex flex-column">
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" id="dithering" v-model="settings.dithering" :disabled="loading" @change="writeSettings" />
                                <label class="form-check-label" for="dithering">Dithering</label>
                            </div>
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" id="usePremiumColors" v-model="settings.usePremiumColors" :disabled="loading" @change="writeSettings" />
                                <label class="form-check-label" for="usePremiumColors">Use premium colors</label>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="grid-item">
            <div class="border rounded p-2 h-100">
                <div class="d-flex flex-column gap-2 h-100">
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" id="buyCharges" v-model="settings.buyCharges" :disabled="loading" @change="writeSettings" />
                        <label class="form-check-label" for="buyCharges">Buy charges</label>
                    </div>
                    <div class="form-group">
                        <label for="buyMaxCharges" class="form-label mb-0">Buy max charges to reach: {{ numberFormat(settings.buyMaxCharges) }}</label>
                        <input type="range" class="form-range" min="0" max="1000" step="10" id="buyMaxCharges" v-model="settings.buyMaxCharges" :disabled="loading" @change="writeSettings" />
                    </div>
                    <div class="form-group">
                        <label for="sleep" class="form-label mb-0">Sleep between loops: {{ numberFormat(settings.sleep) }} seconds</label>
                        <input type="range" class="form-range" min="0" max="1000" step="10" id="sleep" v-model="settings.sleep" :disabled="loading" @change="writeSettings" />
                    </div>
                    <div class="form-group">
                        <label for="requestConcurrent" class="form-label mb-0">Request concurrent: {{ numberFormat(settings.requestConcurrent) }}</label>
                        <input type="range" class="form-range" min="1" max="100" step="1" id="requestConcurrent" v-model="settings.requestConcurrent" :disabled="loading" @change="writeSettings" />
                    </div>

                    <div class="form-group mt-auto">
                        <button v-if="!running" type="button" class="btn btn-primary w-100" @click="start" :disabled="loading || running">
                            <i class="fa-solid fa-play"></i>
                            Start
                        </button>
                        <button v-else type="button" class="btn btn-danger w-100" @click="stop" :disabled="stopping">
                            <i class="fa-solid fa-stop"></i>
                            Stop
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div class="grid-item">
            <div class="border rounded p-2 h-100 overflow-y-auto">
                <div class="d-flex align-items-center justify-content-between gap-2">
                    <button type="button" class="btn btn-sm btn-primary" @click="addUser" :disabled="loading">
                        <i class="fa-solid fa-plus"></i>
                        Add
                    </button>
                    <div class="text-end">
                        <div>Charges: {{ numberFormat(sumBy(settings.users, (x) => x.me?.charges?.count || 0)) }} / {{ numberFormat(sumBy(settings.users, (x) => x.me?.charges?.max || 0)) }}</div>
                        <div>Users: {{ numberFormat(settings.users.length) }}</div>
                    </div>
                </div>
                <table class="table align-middle">
                    <thead>
                        <tr>
                            <th></th>
                            <th>User</th>
                            <th>Charge</th>
                            <th>Droplets</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="user in settings.users" :key="user.id">
                            <td>
                                <input type="checkbox" v-model="user.enabled" :disabled="loading" />
                            </td>
                            <td>{{ user.username }}</td>
                            <td>{{ numberFormat(user.me?.charges?.count || 0) }}/{{ numberFormat(user.me?.charges?.max || 0) }}</td>
                            <td>{{ numberFormat(user.me?.droplets || 0) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div class="grid-item">
            <div class="border rounded p-2 h-100 overflow-y-auto">
                <div v-for="log in logs.toReversed()" class="log-item">{{ log }}</div>
            </div>
        </div>
    </div>
</template>
