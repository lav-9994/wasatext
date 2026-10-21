package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler returns an instance of chi.Mux that handles the APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.Post("/session", rt.doLogin)

	rt.router.Group(func(r chi.Router) {
		r.Use(rt.authMiddleware)

		r.Get("/users", rt.searchUsers)
		r.Put("/users/me/username", rt.setMyUserName)
		r.Put("/users/me/photo", rt.setMyPhoto)
		r.Put("/users/me/bio", rt.setMyBio)
		r.Get("/users/{userId}", rt.getUserProfile)
		r.Get("/users/{userId}/common-groups", rt.getCommonGroups)

		r.Get("/conversations", rt.getMyConversations)
		r.Post("/conversations", rt.startConversation)
		r.Get("/conversations/{conversationId}", rt.getConversation)

		r.Post("/conversations/{conversationId}/messages", rt.sendMessage)
		r.Delete("/conversations/{conversationId}/messages/{messageId}", rt.deleteMessage)
		r.Post("/conversations/{conversationId}/messages/{messageId}/forwards", rt.forwardMessage)
		r.Put("/conversations/{conversationId}/messages/{messageId}/comments/me", rt.commentMessage)
		r.Delete("/conversations/{conversationId}/messages/{messageId}/comments/me", rt.uncommentMessage)

		r.Post("/groups", rt.createGroup)
		r.Post("/groups/{groupId}/members", rt.addToGroup)
		r.Delete("/groups/{groupId}/members/me", rt.leaveGroup)
		r.Put("/groups/{groupId}/name", rt.setGroupName)
		r.Put("/groups/{groupId}/photo", rt.setGroupPhoto)
		r.Put("/groups/{groupId}/description", rt.setGroupDescription)
	})

	// Special routes
	rt.router.Get("/liveness", rt.liveness)

	return rt.router
}
