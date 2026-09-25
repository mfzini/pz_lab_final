<script setup>
import { ref } from 'vue'

const username = ref("");
const password = ref("");

async function wea() {
    const r = await fetch("http://localhost:3000/login", {
        method: "POST",
        mode: "cors",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            Username: username.value,
            Password: password.value
        })
    })
    console.log(await r.json())
}
async function test() {
    const r = await fetch(`http://localhost:3000/person/${username.value}`, {
        method: "GET",
        mode: "cors",
        credentials: "include",
    })
    console.log(await r.json())
}
</script>

<template>

    <h1>Home</h1>
    <form id="loginForm" @submit.prevent="wea">
        <input type="text" name="username" v-model="username">
        <input type="password" name="password" v-model="password">
        <button type="submit" id="submit">Enviar</button>
    </form>
    <form id="loginForm" @submit.prevent="test">
        <button type="submit">Probar</button>
    </form>
</template>

<style lang="css">
h1 {
    color: red;
}
</style>