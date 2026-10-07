<script setup>
    import { ref ,onMounted } from 'vue'
    import { api } from '@/common'
    import { order } from '@/order'
    const Products = ref([])
    const TotalPages = ref(0)
    const CurrentPage = ref(1)
    const cargando = ref(true)

    onMounted(async () => {
        const r = await fetch(`${api}/platos/list`, {mode: "cors"})
        if (!r.ok) { console.log(r) }
        const data = await r.json()
        Products.value = data.Data
        CurrentPage.value = data.CurrentPage
        TotalPages.value = data.TotalPages
        cargando.value = false
    })

    function add(e) {
        const dish_id = e.target.id
        order[dish_id] = document.getElementById(`${dish_id}-amount`).value
        console.log(order)
        for (let prop in order) {
            console.log(`id: ${prop}, amount: ${order[prop]}`)
        }
    }

    
</script>
<template>
    <div v-if="cargando">
        Cargando...
    </div>

    <div v-else>

        <div v-for="(dish) in Products" :key="dish.Id">
            <form @submit.prevent="add" :id="dish.Id">
                <section>
                    <div>{{ dish.Name }}</div>
                    <div>
                        <img :src="`${api}/${dish.Portrait}`" alt="" class="dish-portrait">
                    </div>
                    <div>${{ dish.Price }}</div>
                </section>
                <section>
                    <label for="amount">Cantidad: 
                        <input type="number" name="amount" :id="`${dish.Id}-amount`" value="1">
                    </label>
                </section>
                <button type="submit">Agregar</button>
            </form>
        </div>
        <p>Pagina: {{CurrentPage}} / {{ TotalPages }}</p>
    </div>
</template>


<style>
    .dish-portrait {
        height: 200px;
        width: 200px;
        object-fit: cover;
    }

</style>