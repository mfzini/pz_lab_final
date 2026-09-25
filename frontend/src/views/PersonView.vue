<script setup>
import { ref } from 'vue';
    const api = "http://localhost:3000"
    const user = ref({
        name: "",
        surname: "",
        email: "",
        phone: "",
    })

async function sendRequest() {
    const r = await fetch(`${api}/person`, {
        method: "POST",
        mode: "cors",
        credentials: "include",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(user.value)
    })
    if (!r.ok) {
        console.log(`la request valio callampa ${r.status}`)
    }
    console.log(await r.json())
}
</script>

<template>
    <form id="personForm" @submit.prevent="sendRequest">
        <label for="name">Nombre:</label>
        <input type="text" name="name" id="name" v-model="user.name">
        <label for="surname">Apellido:</label>
        <input type="text" name="surname" id="surname" v-model="user.surname">
        <label for="email">Correo:</label>
        <input type="email" name="email" id="email" v-model="user.email">
        <label for="phone">Teléfono:</label>
        <input type="tel" name="phone" id="phone" v-model="user.phone">
        <button type="submit">Registrar</button>
    </form>
</template>

<style>
    #personForm {
        width: 50%;
        margin: 0 auto;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
    }
    
</style>