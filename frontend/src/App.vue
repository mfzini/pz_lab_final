<script setup>
  
  
</script>

<template>
  <header>
    <router-link to="/">Logo</router-link>
    <nav>
      <router-link to="/about">About</router-link>

    </nav>
  </header>
  <router-view></router-view>
</template>

<style scoped>
header {
  display: flex;
  justify-content: space-between;
  width: 100%;

  nav {
    display: flex;
    gap: 1rem;
  }
}
</style>
