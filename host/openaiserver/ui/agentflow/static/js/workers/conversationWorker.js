/**
 * ConversationWorker - Handles heavy conversation data processing
 * Offloads JSON serialization, conversation management, and data preparation from main thread
 */

class ConversationProcessor {
    constructor() {
        this.conversations = new Map();
        this.compressionThreshold = 50; // messages before compression kicks in
    }

    /**
     * Process and optimize conversation data for storage
     */
    processConversationForStorage(conversationData) {
        try {
            // Clone to avoid modifying original
            const processed = JSON.parse(JSON.stringify(conversationData));

            // Compress large conversations
            if (processed.messages && processed.messages.length > this.compressionThreshold) {
                processed.messages = this.compressMessages(processed.messages);
            }

            // Add metadata
            processed.lastModified = Date.now();
            processed.messageCount = processed.messages ? processed.messages.length : 0;
            processed.size = JSON.stringify(processed).length;

            return {
                success: true,
                data: processed,
                metadata: {
                    compressed: processed.messages.length > this.compressionThreshold,
                    originalSize: JSON.stringify(conversationData).length,
                    processedSize: processed.size
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
     * Compress conversation messages for storage efficiency
     */
    compressMessages(messages) {
        return messages.map(msg => {
            const compressed = {
                role: msg.role,
                content: msg.content,
                timestamp: msg.timestamp
            };

            // Only include attachments if they exist
            if (msg.attachments && msg.attachments.length > 0) {
                compressed.attachments = msg.attachments.map(att => {
                    // For large data URLs, store reference instead
                    if (att.type === 'image_url' && att.image_url?.url?.length > 100000) {
                        return {
                            ...att,
                            image_url: {
                                url: '[COMPRESSED]',
                                originalSize: att.image_url.url.length
                            }
                        };
                    }
                    return att;
                });
            }

            return compressed;
        });
    }

    /**
     * Prepare conversation data for API requests
     */
    prepareForAPI(conversation, systemPrompt, selectedTools) {
        try {
            const messages = [];

            // Add system prompt if provided
            if (systemPrompt && systemPrompt.trim()) {
                messages.push({
                    role: 'system',
                    content: systemPrompt
                });
            }

            // Process conversation messages
            conversation.messages.forEach(msg => {
                if (msg.role === 'user' || msg.role === 'assistant') {
                    const apiMessage = {
                        role: msg.role,
                        content: msg.content
                    };

                    // Add attachments for user messages
                    if (msg.role === 'user' && msg.attachments && msg.attachments.length > 0) {
                        // Convert to OpenAI format
                        const content = [{ type: 'text', text: msg.content }];

                        msg.attachments.forEach(att => {
                            if (att.type === 'image_url') {
                                content.push({
                                    type: 'image_url',
                                    image_url: att.image_url
                                });
                            }
                        });

                        apiMessage.content = content;
                    }

                    messages.push(apiMessage);
                }
            });

            return {
                success: true,
                data: {
                    messages,
                    tools: selectedTools.length > 0 ? selectedTools : undefined,
                    messageCount: messages.length
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
     * Calculate conversation statistics
     */
    calculateStats(conversations) {
        const stats = {
            totalConversations: Object.keys(conversations).length,
            totalMessages: 0,
            totalSize: 0,
            oldestConversation: null,
            newestConversation: null,
            largestConversation: null
        };

        let oldestTime = Infinity;
        let newestTime = 0;
        let largestSize = 0;

        Object.entries(conversations).forEach(([id, conv]) => {
            const messageCount = conv.messages ? conv.messages.length : 0;
            stats.totalMessages += messageCount;

            const convData = JSON.stringify(conv);
            const size = convData.length;
            stats.totalSize += size;

            const timestamp = conv.timestamp || 0;
            if (timestamp < oldestTime) {
                oldestTime = timestamp;
                stats.oldestConversation = { id, timestamp, messageCount };
            }

            if (timestamp > newestTime) {
                newestTime = timestamp;
                stats.newestConversation = { id, timestamp, messageCount };
            }

            if (size > largestSize) {
                largestSize = size;
                stats.largestConversation = { id, size, messageCount };
            }
        });

        return stats;
    }

    /**
     * Clean up old conversations based on criteria
     */
    suggestCleanup(conversations, maxSize = 5 * 1024 * 1024) { // 5MB default
        const stats = this.calculateStats(conversations);

        if (stats.totalSize < maxSize) {
            return { needsCleanup: false, stats };
        }

        // Sort conversations by last access time and size
        const sortedConvs = Object.entries(conversations)
            .map(([id, conv]) => ({
                id,
                timestamp: conv.timestamp || 0,
                size: JSON.stringify(conv).length,
                messageCount: conv.messages ? conv.messages.length : 0
            }))
            .sort((a, b) => a.timestamp - b.timestamp); // oldest first

        const toRemove = [];
        let currentSize = stats.totalSize;

        for (const conv of sortedConvs) {
            if (currentSize <= maxSize) break;
            toRemove.push(conv.id);
            currentSize -= conv.size;
        }

        return {
            needsCleanup: true,
            stats,
            suggestedRemovals: toRemove,
            estimatedSizeAfterCleanup: currentSize
        };
    }
}

// Worker event handling
const processor = new ConversationProcessor();

self.onmessage = function(e) {
    const { type, data, id } = e.data;

    try {
        let result;

        switch (type) {
            case 'processForStorage':
                result = processor.processConversationForStorage(data);
                break;

            case 'prepareForAPI':
                result = processor.prepareForAPI(data.conversation, data.systemPrompt, data.selectedTools);
                break;

            case 'calculateStats':
                result = { success: true, data: processor.calculateStats(data) };
                break;

            case 'suggestCleanup':
                result = { success: true, data: processor.suggestCleanup(data.conversations, data.maxSize) };
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