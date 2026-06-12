<script setup lang="ts">
import LegalPageShell from '~/components/legal/LegalPageShell.vue'

definePageMeta({ layout: false })

useHead({ title: "Conditions générales d'utilisation — SMO" })
</script>

<template>
  <LegalPageShell title="Conditions générales d'utilisation" updated-at="12 juin 2026">
    <section>
      <h2>Le service</h2>
      <p>
        SMO permet d'organiser des matchs entre connaissances : un organisateur
        crée un groupe et des matchs, les joueurs confirment leur présence par
        lien d'invitation, puis notent leurs coéquipiers. Le service repose sur
        un modèle de confiance entre membres d'un même groupe.
      </p>
    </section>

    <section>
      <h2>Compte organisateur</h2>
      <p>
        L'organisateur est responsable des prénoms qu'il saisit et s'engage à
        informer les joueurs concernés de l'utilisation de SMO pour son groupe.
      </p>
    </section>

    <section>
      <h2>Liens d'invitation</h2>
      <p>
        Votre lien d'invitation est un lien personnel : il vaut identification,
        ne le partagez pas en dehors de votre groupe. Réclamer un prénom sur un
        lien de match est définitif pour ce match. Toute utilisation du lien
        d'autrui est interdite.
      </p>
    </section>

    <section>
      <h2>Disponibilité et garantie</h2>
      <p>
        SMO est un projet en développement, fourni en l'état, sans garantie de
        disponibilité ni d'absence d'erreur (licence MIT). L'éditeur peut
        suspendre ou faire évoluer le service à tout moment.
      </p>
    </section>

    <section>
      <h2>Droit applicable</h2>
      <p>
        Les présentes conditions sont régies par le droit français.
      </p>
    </section>
  </LegalPageShell>
</template>
