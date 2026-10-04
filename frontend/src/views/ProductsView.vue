<script setup>
import { ref } from 'vue'

    const api = "http://localhost:3000"
    const dish = ref({
        name: "",
        price: 0,
    })
    function renderPreview(e) {
        document.getElementById('preview').src = URL.createObjectURL(e.target.files[0])
        document.getElementById('previewContainer').style.display = 'flex'
    }

    async function submitDish() {
        const picture = document.getElementById('picture')
        const formData = new FormData()
        console.log(JSON.stringify(dish.value))
        formData.append("dish", JSON.stringify(dish.value))
        formData.append('picture', picture.files[0])
        const r = await fetch(`${api}/platos`, {
            method: "POST",
            mode: "cors",
            body: formData,
        })
        console.log(r)
        console.log(await r.json())
    }

</script>

<template>
    <div id="container">
        <form @submit.prevent="submitDish">
            <label for="name">Nombre:</label>
            <input type="text" name="name" id="name" v-model="dish.name">
            <label for="price">Precio:</label>
            <input type="number" name="price" id="price" v-model="dish.price">
            <label for="picture">Foto:</label>
            <input type="file" name="picture" id="picture" @change="renderPreview" accept=".png, .jpg, .jpeg">
            <input type="submit" value="Enviar">
        </form>
        <div id="previewContainer">
            <img src="" alt="" id="preview">
        </div>
    </div>
</template>

<style>
    #container {
        width: 100%;
        display: flex;
        justify-content: center;
    }
    form {
        display: flex;
        justify-content: center;
        align-items: center;
        flex-direction: column;
        gap: 3rem;
    }
    #previewContainer {
        display: none;
        margin: auto;
        height: 200px;
        width: 200px;
        #preview {
            height: 200px;
        width: 200px;
        object-fit: cover;
        }
    }

</style>