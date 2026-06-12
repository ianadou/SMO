<script setup lang="ts">
import LegalPageShell from '~/components/legal/LegalPageShell.vue'

definePageMeta({ layout: false })

useHead({ title: 'Politique de confidentialité — SMO' })
</script>

<template>
  <LegalPageShell title="Politique de confidentialité" updated-at="12 juin 2026">
    <section>
      <h2>Responsable de traitement</h2>
      <p>
        Eddin ADOU. Pour toute question ou demande relative à vos données :
        <a href="mailto:ianadou807@gmail.com">ianadou807@gmail.com</a>.
      </p>
    </section>

    <section>
      <h2>Données traitées</h2>
      <ul>
        <li>
          Compte organisateur : adresse email, nom affiché et mot de passe.
          Le mot de passe est haché (bcrypt) et n'est jamais stocké en clair.
        </li>
        <li>
          Joueurs : prénom uniquement, saisi par l'organisateur ou par le
          joueur lui-même via un lien d'invitation. Les joueurs n'ont aucun
          compte — ni email, ni numéro de téléphone.
        </li>
        <li>
          Liens d'invitation et de match : jetons aléatoires, stockés
          uniquement sous forme hachée (SHA-256).
        </li>
        <li>
          Votes : note de 1 à 5 attribuée à ses coéquipiers après un match.
          Vos coéquipiers ne voient jamais qui les a notés.
          L'organisateur du groupe peut consulter le détail des votes de ses
          matchs. Les notes alimentent le classement des joueurs du groupe.
        </li>
        <li>
          Journaux techniques : adresse IP et identifiant de requête,
          conservés à des fins de sécurité et de diagnostic. Les jetons n'y
          figurent jamais en clair.
        </li>
      </ul>
    </section>

    <section>
      <h2>Stockage local, sans cookie</h2>
      <p>
        SMO n'utilise aucun cookie. Le navigateur conserve dans son
        localStorage : le jeton de session de l'organisateur connecté, et,
        côté joueur, le lien personnel d'un match rejoint (pour le retrouver
        au retour). Se déconnecter ou effacer les données du navigateur les
        supprime.
      </p>
    </section>

    <section>
      <h2>Finalités et base légale</h2>
      <p>
        Les données servent exclusivement à organiser les matchs : inviter,
        confirmer les présences, composer les équipes et établir le classement
        du groupe. Base légale : l'exécution du service demandé, et l'intérêt
        légitime pour la sécurité. Aucune publicité, aucune mesure d'audience,
        aucun partage à des tiers — à l'exception d'une notification Discord
        si l'organisateur a configuré son propre webhook de groupe.
      </p>
    </section>

    <section>
      <h2>Vos droits</h2>
      <p>
        Conformément au RGPD, vous disposez d'un droit d'accès, de
        rectification, d'effacement et d'opposition. Écrivez à
        <a href="mailto:ianadou807@gmail.com">ianadou807@gmail.com</a> — y
        compris si vous êtes un joueur sans compte dont un prénom a été saisi.
        Vous pouvez également saisir la CNIL (cnil.fr).
      </p>
    </section>
  </LegalPageShell>
</template>
