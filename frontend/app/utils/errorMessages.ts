import { ApiError } from '~/composables/useApi'

interface FriendlyMessage {
  title: string
  message?: string
}

const BACKEND_TO_FR: Record<string, FriendlyMessage> = {
  'group not found': { title: 'Groupe introuvable', message: 'Ce groupe n\'existe pas ou n\'est plus accessible.' },
  'match not found': { title: 'Match introuvable', message: 'Ce match n\'existe pas ou a été supprimé.' },
  'player not found': { title: 'Joueur introuvable' },
  'invitation not found': { title: 'Invitation introuvable', message: 'Le lien d\'invitation n\'est plus valide.' },
  'vote not found': { title: 'Vote introuvable' },
  'organizer not found': { title: 'Organisateur introuvable' },

  'invalid id': { title: 'Identifiant invalide' },
  'invalid name': { title: 'Nom invalide', message: 'Le nom doit faire entre 1 et 100 caractères.' },
  'invalid date': { title: 'Date invalide', message: 'La date doit être au format valide et dans le futur.' },
  'invalid status': { title: 'Statut invalide pour cette action' },
  'invalid parameter': { title: 'Paramètre invalide' },
  'referenced entity does not exist': { title: 'Référence invalide', message: 'Un élément référencé n\'existe pas.' },
  'invalid email': { title: 'Email invalide', message: 'Vérifie l\'orthographe de ton adresse email.' },
  'invalid password': { title: 'Mot de passe invalide', message: 'Le mot de passe doit faire au moins 12 caractères.' },
  'invalid webhook url': { title: 'URL de webhook invalide' },
  'invalid invitation response': { title: 'Réponse d\'invitation invalide' },
  'invalid score': { title: 'Score invalide' },
  'invalid match score': { title: 'Score du match invalide' },

  'operation not allowed in current state': {
    title: 'Action impossible',
    message: 'Le match n\'est pas dans un état qui autorise cette action.',
  },
  'invalid team assignment': {
    title: 'Génération impossible',
    message: 'Il n\'y a pas assez de joueurs confirmés pour former deux équipes.',
  },
  'cannot vote for yourself': { title: 'Vote refusé', message: 'Tu ne peux pas voter pour toi-même.' },
  'match is full': { title: 'Match complet', message: 'Le nombre maximum de joueurs est atteint.' },
  'team is full': { title: 'Équipe complète' },
  'player is not in this match': { title: 'Joueur absent', message: 'Ce joueur ne participe pas à ce match.' },
  'invitation expired': { title: 'Invitation expirée', message: 'Le lien d\'invitation n\'est plus actif.' },
  'invitation can no longer be changed': {
    title: 'Réponse figée',
    message: 'La réponse à cette invitation ne peut plus être modifiée.',
  },
  'already voted for this player in this match': {
    title: 'Vote déjà enregistré',
    message: 'Tu as déjà voté pour ce joueur sur ce match.',
  },
  'match is not completed': { title: 'Match non terminé', message: 'Cette action requiert un match terminé.' },
  'email already exists': { title: 'Email déjà utilisé', message: 'Un compte existe déjà avec cette adresse.' },

  'rate limit exceeded': {
    title: 'Trop de requêtes',
    message: 'Patiente quelques secondes avant de réessayer.',
  },
  'unauthorized': { title: 'Non autorisé', message: 'Reconnecte-toi pour continuer.' },
  'forbidden': { title: 'Accès refusé', message: 'Tu n\'as pas le droit d\'effectuer cette action.' },
  'internal server error': {
    title: 'Erreur serveur',
    message: 'Une erreur inattendue est survenue. Réessaie dans un instant.',
  },
}

const STATUS_FALLBACKS: Record<number, FriendlyMessage> = {
  400: { title: 'Requête invalide', message: 'Les données envoyées ne sont pas valides.' },
  401: { title: 'Non autorisé', message: 'Reconnecte-toi pour continuer.' },
  403: { title: 'Accès refusé', message: 'Tu n\'as pas le droit d\'effectuer cette action.' },
  404: { title: 'Introuvable', message: 'La ressource demandée n\'existe pas.' },
  409: { title: 'Conflit', message: 'L\'action est en conflit avec l\'état actuel.' },
  410: { title: 'Expiré', message: 'La ressource n\'est plus accessible.' },
  429: { title: 'Trop de requêtes', message: 'Patiente quelques secondes avant de réessayer.' },
  500: { title: 'Erreur serveur', message: 'Une erreur inattendue est survenue.' },
}

export function toFriendlyError(err: unknown): FriendlyMessage {
  if (err instanceof ApiError) {
    const mapped = BACKEND_TO_FR[err.publicMessage]
    if (mapped) return mapped
    const fallback = STATUS_FALLBACKS[err.status]
    if (fallback) return fallback
    return { title: 'Erreur', message: err.publicMessage }
  }
  if (err instanceof Error) {
    return { title: 'Erreur', message: err.message }
  }
  return { title: 'Erreur', message: 'Une erreur inattendue est survenue.' }
}
