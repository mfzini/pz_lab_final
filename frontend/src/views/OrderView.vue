<script setup>

    const api = "http://localhost:3000"
    const r = await fetch(`${api}/platos/list`, {mode: "cors"})
    if (!r.ok) { console.log(r) }

    const {Data, TotalPages, CurrentPage } = await r.json()
    console.log(Data)
    console.log(TotalPages)
</script>

<template>

    <div v-for="(dish) in Data" :key="dish.Id">
        <div :id="dish.Id">
            <div>{{ dish.Name }}</div>
            <div>
                <img :src="`${api}/${dish.Portrait}`" alt="" class="dish-portrait">
            </div>
            <div>${{ dish.Price }}</div>
        </div>
    </div>
    <p>Pagina: {{CurrentPage}} / {{ TotalPages }}</p>
</template>


<style>
    .dish-portrait {
        height: 200px;
        width: 200px;
        object-fit: cover;
    }

</style>