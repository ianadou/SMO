<script setup lang="ts">
import LegalPageShell from '~/components/legal/LegalPageShell.vue'

definePageMeta({ layout: false })

useHead({ title: 'Mentions légales — SMO' })
</script>

<template>
  <LegalPageShell title="Mentions légales" updated-at="12 juin 2026">
    <section>
      <h2>Éditeur</h2>
      <p>
        SMO est édité par Eddin ADOU, qui en est également le directeur de la
        publication.
      </p>
      <p>
        Contact : <a href="mailto:ianadou807@gmail.com">ianadou807@gmail.com</a>
      </p>
    </section>

    <section>
      <h2>Hébergement</h2>
      <p>
        L'application est en cours de développement et n'est pas encore déployée
        publiquement. Le déploiement est prévu chez OVHcloud (OVH SAS, 2 rue
        Kellermann, 59100 Roubaix, France). Cette section sera complétée à la
        mise en ligne.
      </p>
    </section>

    <section>
      <h2>Propriété intellectuelle</h2>
      <p>
        Le code source de SMO est publié sous licence MIT. Les contenus saisis
        dans l'application (noms de groupes, matchs, prénoms) restent la
        propriété de leurs auteurs.
      </p>
    </section>
  </LegalPageShell>
</template>
