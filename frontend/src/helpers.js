import Swal from 'sweetalert2';

export function generateId() {
    return Date.now() + Math.random().toString(36).substring(2, 9);
}

export function sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function alert(title, text, icon = 'success') {
    await Swal.fire({
        title,
        text,
        icon,
        theme: 'dark',
        backdrop: `rgba(0,0,123,0.4) url("https://sweetalert2.github.io/images/nyan-cat.gif") left top no-repeat`,
    });
}

export async function confirm(title, text, icon = 'question') {
    const result = await Swal.fire({
        title,
        text,
        icon,
        theme: 'dark',
        backdrop: `rgba(0,0,123,0.4) url("https://sweetalert2.github.io/images/nyan-cat.gif") left top no-repeat`,
        showCancelButton: true,
        focusCancel: true,
    });
    return result.isConfirmed;
}

export async function input(title) {
    const result = await Swal.fire({
        title,
        input: 'text',
        theme: 'dark',
        backdrop: `rgba(0,0,123,0.4) url("https://sweetalert2.github.io/images/nyan-cat.gif") left top no-repeat`,
        showCancelButton: true,
        focusCancel: true,
    });

    return result.isConfirmed ? result.value : null;
}

export async function imageDataFromBuffer(buffer) {
    const blob = new Blob([buffer], { type: 'image/png' });
    const bitmap = await createImageBitmap(blob);

    const canvas = document.createElement('canvas');
    canvas.width = bitmap.width;
    canvas.height = bitmap.height;

    const ctx = canvas.getContext('2d');
    ctx.drawImage(bitmap, 0, 0);

    const data = ctx.getImageData(0, 0, canvas.width, canvas.height);
    return data;
}

export function numberFormat(value) {
    value = parseInt(value);
    return value.toLocaleString('en-US');
}

export function isUnsignedInteger(value) {
    return /^\d+$/.test(value);
}

export function randomstring(length = 8) {
    return Math.random().toString(36).substring(2, 2 + length);
}
