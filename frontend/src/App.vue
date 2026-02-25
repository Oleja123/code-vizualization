<template>
  <div id="app">
    <header class="app-header">
      <h1>Визуализация кода</h1>
      <nav>
        <button class="nav-button active">Визуализация</button>
      </nav>
    </header>
    <main class="app-main">
      <VisualizationView />
    </main>
  </div>
</template>

<script>
import VisualizationView from './views/VisualizationView.vue'

export default {
  name: 'App',
  components: {
    VisualizationView
  }
}
</script>

<style scoped>
.app-header {
  background-color: #2c3e50;
  color: white;
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.app-header h1 {
  font-size: 1.5rem;
  font-weight: 600;
}

.nav-button {
  background-color: #34495e;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  transition: background-color 0.2s;
}

.nav-button:hover {
  background-color: #415a77;
}

.nav-button.active {
  background-color: #3498db;
}

.app-main {
  flex: 1;
  overflow: hidden;
}
</style>
