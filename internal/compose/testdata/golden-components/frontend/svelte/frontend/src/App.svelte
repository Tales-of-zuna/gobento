<script lang="ts">
  // The example page: lists the items of the API and adds new ones.
  import { onMount } from 'svelte'
  import { createItem, errorMessage, listItems, type Item } from './api'

  let items = $state<Item[]>([])
  let name = $state('')
  let error = $state('')

  onMount(async () => {
    try {
      items = await listItems()
    } catch (err) {
      error = errorMessage(err)
    }
  })

  async function add(event: SubmitEvent) {
    event.preventDefault()
    try {
      const item = await createItem(name)
      items = [...items, item]
      name = ''
      error = ''
    } catch (err) {
      error = errorMessage(err)
    }
  }
</script>

<main>
  <h1>shop</h1>
  <form onsubmit={add}>
    <label for="name">Name</label>
    <input id="name" name="name" autocomplete="off" bind:value={name} />
    <button type="submit">Add</button>
  </form>
  {#if error}
    <p role="alert">{error}</p>
  {/if}
  {#if items.length === 0}
    <p class="empty">No items yet.</p>
  {:else}
    <ul>
      {#each items as item (item.id)}
        <li>{item.name}</li>
      {/each}
    </ul>
  {/if}
</main>
