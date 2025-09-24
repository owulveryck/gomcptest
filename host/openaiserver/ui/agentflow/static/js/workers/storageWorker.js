/**
 * StorageWorker - Handles storage data processing in background
 * Processes and optimizes data for storage without blocking main thread
 * Note: Actual storage operations happen on main thread via message passing
 */

class StorageManager {
    constructor() {
        this.maxFailures = 3;
    }

    /**
     * Safely truncate content - handles both string and array/object types
     */
    truncateContent(content, maxLength) {
        if (!content) return '';

        if (typeof content === 'string') {
            return content.substring(0, maxLength);
        }

        // For array/object content (multipart messages), convert to string first
        if (Array.isArray(content) || typeof content === 'object') {
            const stringified = JSON.stringify(content);
            return stringified.substring(0, maxLength);
        }

        return String(content).substring(0, maxLength);
    }

    /**
     * Process conversations for storage optimization
     */
    processConversationsForStorage(conversations) {
        try {
            const processed = JSON.parse(JSON.stringify(conversations));
            const stats = this.calculateStorageUsage(processed);

            return {
                success: true,
                data: {
                    conversations: processed,
                    stats: stats,
                    optimized: false
                }
            };
        } catch (error) {
            return {
                success: false,
                error: error.message
            };
        }
    }

    /**
     * Create reduced version of conversations for storage quota issues
     */
    createReducedConversations(conversations) {
        const reduced = {};

        Object.entries(conversations).forEach(([id, conv]) => {
            reduced[id] = {
                title: conv.title,
                timestamp: conv.timestamp,
                messages: conv.messages ? conv.messages.map(msg => ({
                    role: msg.role,
                    content: this.truncateContent(msg.content, 1000), // Safely limit content
                    timestamp: msg.timestamp,
                    // Skip large attachments
                    attachments: msg.attachments ? msg.attachments.filter(att =>
                        !att.image_url?.url || att.image_url.url.length < 10000
                    ) : []
                })) : []
            };
        });

        return {
            success: true,
            data: reduced
        };
    }

    /**
     * Create emergency backup data
     */
    createEmergencyData(conversations) {
        const emergency = {};
        const entries = Object.entries(conversations);

        // Only keep most recent 5 conversations
        const recent = entries
            .sort(([, a], [, b]) => (b.timestamp || 0) - (a.timestamp || 0))
            .slice(0, 5);

        recent.forEach(([id, conv]) => {
            emergency[id] = {
                title: conv.title,
                timestamp: conv.timestamp,
                messages: conv.messages ? conv.messages.slice(-10).map(msg => ({ // Only last 10 messages
                    role: msg.role,
                    content: this.truncateContent(msg.content, 500), // Heavily truncated
                    timestamp: msg.timestamp
                })) : []
            };
        });

        return {
            success: true,
            data: emergency
        };
    }

    /**
     * Clean up old conversations based on criteria
     */
    suggestCleanup(conversations, maxCount = 50, maxAge = 30 * 24 * 60 * 60 * 1000) { // 30 days
        const now = Date.now();
        const entries = Object.entries(conversations);
        const toRemove = [];

        // Remove conversations older than maxAge
        entries.forEach(([id, conv]) => {
            const age = now - (conv.timestamp || 0);
            if (age > maxAge) {
                toRemove.push({ id, reason: 'age', age });
            }
        });

        // If still too many, remove oldest ones
        if (entries.length - toRemove.length > maxCount) {
            const remaining = entries.filter(([id]) =>
                !toRemove.find(item => item.id === id)
            );

            const oldestFirst = remaining.sort(([, a], [, b]) =>
                (a.timestamp || 0) - (b.timestamp || 0)
            );

            const excessCount = remaining.length - maxCount;
            oldestFirst.slice(0, excessCount).forEach(([id]) => {
                toRemove.push({ id, reason: 'count' });
            });
        }

        return {
            success: true,
            data: {
                removedItems: toRemove,
                remainingCount: entries.length - toRemove.length
            }
        };
    }

    /**
     * Calculate storage usage
     */
    calculateStorageUsage(conversations) {
        const totalSize = JSON.stringify(conversations).length;
        const conversationCount = Object.keys(conversations).length;
        const messageCount = Object.values(conversations).reduce((total, conv) =>
            total + (conv.messages ? conv.messages.length : 0), 0
        );

        return {
            totalSize,
            conversationCount,
            messageCount,
            averageConversationSize: conversationCount > 0 ? Math.round(totalSize / conversationCount) : 0,
            averageMessageSize: messageCount > 0 ? Math.round(totalSize / messageCount) : 0
        };
    }

    /**
     * Optimize conversations data for better performance
     */
    optimizeConversationsData(conversations) {
        try {
            // Deep clone to avoid modifying original
            const optimized = JSON.parse(JSON.stringify(conversations));

            // Remove redundant data and optimize structure
            Object.values(optimized).forEach(conversation => {
                if (conversation.messages) {
                    conversation.messages.forEach(message => {
                        // Remove empty attachments arrays
                        if (message.attachments && message.attachments.length === 0) {
                            delete message.attachments;
                        }

                        // Trim whitespace from content
                        if (message.content && typeof message.content === 'string') {
                            message.content = message.content.trim();
                        }
                    });
                }
            });

            return {
                success: true,
                data: {
                    original: conversations,
                    optimized: optimized,
                    sizeDifference: JSON.stringify(conversations).length - JSON.stringify(optimized).length
                }
            };
        } catch (error) {
            return {
                success: false,
                error: error.message
            };
        }
    }
}

// Worker event handling
const storageManager = new StorageManager();

self.onmessage = async function(e) {
    const { type, data, id } = e.data;

    try {
        let result;

        switch (type) {
            case 'processConversationsForStorage':
                result = storageManager.processConversationsForStorage(data);
                break;

            case 'createReducedConversations':
                result = storageManager.createReducedConversations(data);
                break;

            case 'createEmergencyData':
                result = storageManager.createEmergencyData(data);
                break;

            case 'suggestCleanup':
                result = storageManager.suggestCleanup(
                    data.conversations,
                    data.maxCount,
                    data.maxAge
                );
                break;

            case 'calculateStorageUsage':
                result = { success: true, data: storageManager.calculateStorageUsage(data) };
                break;

            case 'optimizeConversationsData':
                result = storageManager.optimizeConversationsData(data);
                break;

            default:
                result = { success: false, error: `Unknown operation: ${type}` };
        }

        self.postMessage({ id, type, ...result });
    } catch (error) {
        self.postMessage({
            id,
            type,
            success: false,
            error: error.message
        });
    }
};

// Handle worker initialization
self.postMessage({ type: 'ready' });