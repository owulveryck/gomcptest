/**
 * MessageWorker - Handles message content processing and preparation
 * Processes markdown, attachments, and content formatting without blocking main thread
 */

class MessageProcessor {
    constructor() {
        this.markedOptions = {
            breaks: true,
            gfm: true,
            sanitize: false
        };
    }

    /**
     * Process message content for display
     */
    processMessageContent(message) {
        try {
            const processed = {
                ...message,
                processedContent: message.content,
                hasCode: false,
                hasImages: false,
                hasAudio: false,
                attachmentCount: 0
            };

            // Check for code blocks
            if (message.content && typeof message.content === 'string') {
                processed.hasCode = /```[\s\S]*?```|`[^`]+`/.test(message.content);
            }

            // Process attachments
            if (message.attachments && message.attachments.length > 0) {
                processed.attachmentCount = message.attachments.length;

                message.attachments.forEach(att => {
                    if (att.type === 'image_url') {
                        processed.hasImages = true;
                    } else if (att.type === 'audio') {
                        processed.hasAudio = true;
                    }
                });
            }

            return {
                success: true,
                data: processed
            };
        } catch (error) {
            return {
                success: false,
                error: error.message
            };
        }
    }

    /**
     * Process messages for search functionality
     */
    processMessagesForSearch(messages, searchTerm) {
        if (!searchTerm || searchTerm.trim() === '') {
            return {
                success: true,
                data: {
                    results: [],
                    totalMatches: 0
                }
            };
        }

        try {
            const term = searchTerm.toLowerCase().trim();
            const results = [];

            messages.forEach((message, index) => {
                if (message.content && typeof message.content === 'string') {
                    const content = message.content.toLowerCase();

                    if (content.includes(term)) {
                        // Find all matches in the content
                        const matches = [];
                        let position = content.indexOf(term);

                        while (position !== -1) {
                            // Get context around the match
                            const start = Math.max(0, position - 50);
                            const end = Math.min(content.length, position + term.length + 50);
                            const context = message.content.substring(start, end);

                            matches.push({
                                position,
                                context,
                                highlighted: this.highlightSearchTerm(context, searchTerm)
                            });

                            position = content.indexOf(term, position + 1);
                        }

                        if (matches.length > 0) {
                            results.push({
                                messageIndex: index,
                                role: message.role,
                                timestamp: message.timestamp,
                                matches,
                                matchCount: matches.length
                            });
                        }
                    }
                }
            });

            const totalMatches = results.reduce((sum, result) => sum + result.matchCount, 0);

            return {
                success: true,
                data: {
                    results,
                    totalMatches,
                    searchTerm
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
     * Highlight search terms in text
     */
    highlightSearchTerm(text, searchTerm) {
        if (!searchTerm || !text) return text;

        const regex = new RegExp(`(${this.escapeRegExp(searchTerm)})`, 'gi');
        return text.replace(regex, '<mark>$1</mark>');
    }

    /**
     * Escape special regex characters
     */
    escapeRegExp(string) {
        return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }

    /**
     * Process conversation for export
     */
    processConversationForExport(conversation, format = 'json') {
        try {
            switch (format.toLowerCase()) {
                case 'json':
                    return this.exportAsJSON(conversation);
                case 'markdown':
                    return this.exportAsMarkdown(conversation);
                case 'text':
                    return this.exportAsText(conversation);
                case 'html':
                    return this.exportAsHTML(conversation);
                default:
                    throw new Error(`Unsupported export format: ${format}`);
            }
        } catch (error) {
            return {
                success: false,
                error: error.message
            };
        }
    }

    /**
     * Export conversation as JSON
     */
    exportAsJSON(conversation) {
        const exportData = {
            title: conversation.title || 'Untitled Conversation',
            timestamp: conversation.timestamp,
            exportedAt: new Date().toISOString(),
            messageCount: conversation.messages ? conversation.messages.length : 0,
            messages: conversation.messages || []
        };

        return {
            success: true,
            data: {
                content: JSON.stringify(exportData, null, 2),
                filename: `${this.sanitizeFilename(exportData.title)}_${new Date().toISOString().split('T')[0]}.json`,
                mimeType: 'application/json'
            }
        };
    }

    /**
     * Export conversation as Markdown
     */
    exportAsMarkdown(conversation) {
        const title = conversation.title || 'Untitled Conversation';
        const date = conversation.timestamp ? new Date(conversation.timestamp).toLocaleString() : 'Unknown';

        let markdown = `# ${title}\n\n`;
        markdown += `**Conversation Date:** ${date}\n`;
        markdown += `**Exported:** ${new Date().toLocaleString()}\n\n`;
        markdown += `---\n\n`;

        if (conversation.messages) {
            conversation.messages.forEach((message, index) => {
                const role = message.role === 'user' ? '👤 User' : '🤖 Assistant';
                const timestamp = message.timestamp ?
                    new Date(message.timestamp).toLocaleTimeString() : '';

                markdown += `## ${role}${timestamp ? ` (${timestamp})` : ''}\n\n`;

                if (message.content) {
                    markdown += `${message.content}\n\n`;
                }

                if (message.attachments && message.attachments.length > 0) {
                    markdown += `**Attachments:** ${message.attachments.length}\n\n`;
                }

                if (index < conversation.messages.length - 1) {
                    markdown += `---\n\n`;
                }
            });
        }

        return {
            success: true,
            data: {
                content: markdown,
                filename: `${this.sanitizeFilename(title)}_${new Date().toISOString().split('T')[0]}.md`,
                mimeType: 'text/markdown'
            }
        };
    }

    /**
     * Export conversation as plain text
     */
    exportAsText(conversation) {
        const title = conversation.title || 'Untitled Conversation';
        const date = conversation.timestamp ? new Date(conversation.timestamp).toLocaleString() : 'Unknown';

        let text = `${title}\n`;
        text += `${'='.repeat(title.length)}\n\n`;
        text += `Conversation Date: ${date}\n`;
        text += `Exported: ${new Date().toLocaleString()}\n\n`;

        if (conversation.messages) {
            conversation.messages.forEach((message, index) => {
                const role = message.role === 'user' ? 'USER' : 'ASSISTANT';
                const timestamp = message.timestamp ?
                    new Date(message.timestamp).toLocaleTimeString() : '';

                text += `[${role}]${timestamp ? ` ${timestamp}` : ''}\n`;
                text += `${'-'.repeat(50)}\n`;

                if (message.content) {
                    text += `${message.content}\n`;
                }

                if (message.attachments && message.attachments.length > 0) {
                    text += `\n[Attachments: ${message.attachments.length}]\n`;
                }

                text += `\n`;
            });
        }

        return {
            success: true,
            data: {
                content: text,
                filename: `${this.sanitizeFilename(title)}_${new Date().toISOString().split('T')[0]}.txt`,
                mimeType: 'text/plain'
            }
        };
    }

    /**
     * Export conversation as HTML
     */
    exportAsHTML(conversation) {
        const title = conversation.title || 'Untitled Conversation';
        const date = conversation.timestamp ? new Date(conversation.timestamp).toLocaleString() : 'Unknown';

        let html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${this.escapeHtml(title)}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; line-height: 1.6; }
        .header { border-bottom: 2px solid #eee; padding-bottom: 20px; margin-bottom: 30px; }
        .message { margin-bottom: 30px; padding: 20px; border-radius: 8px; }
        .user { background-color: #f0f9ff; border-left: 4px solid #0ea5e9; }
        .assistant { background-color: #f9fafb; border-left: 4px solid #6b7280; }
        .role { font-weight: bold; margin-bottom: 10px; color: #374151; }
        .timestamp { font-size: 0.9em; color: #6b7280; }
        .content { margin-top: 10px; }
        .attachments { margin-top: 10px; font-style: italic; color: #6b7280; }
        pre { background: #f3f4f6; padding: 15px; border-radius: 6px; overflow-x: auto; }
        code { background: #f3f4f6; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>${this.escapeHtml(title)}</h1>
        <p><strong>Conversation Date:</strong> ${this.escapeHtml(date)}</p>
        <p><strong>Exported:</strong> ${this.escapeHtml(new Date().toLocaleString())}</p>
    </div>
`;

        if (conversation.messages) {
            conversation.messages.forEach(message => {
                const role = message.role === 'user' ? 'User' : 'Assistant';
                const timestamp = message.timestamp ?
                    new Date(message.timestamp).toLocaleTimeString() : '';

                html += `    <div class="message ${message.role}">
        <div class="role">${role}${timestamp ? ` <span class="timestamp">(${timestamp})</span>` : ''}</div>
        <div class="content">${this.escapeHtml(message.content || '')}</div>`;

                if (message.attachments && message.attachments.length > 0) {
                    html += `        <div class="attachments">📎 ${message.attachments.length} attachment(s)</div>`;
                }

                html += `    </div>\n`;
            });
        }

        html += `</body>
</html>`;

        return {
            success: true,
            data: {
                content: html,
                filename: `${this.sanitizeFilename(title)}_${new Date().toISOString().split('T')[0]}.html`,
                mimeType: 'text/html'
            }
        };
    }

    /**
     * Sanitize filename for download
     */
    sanitizeFilename(filename) {
        return filename
            .replace(/[^a-z0-9]/gi, '_')
            .replace(/_+/g, '_')
            .replace(/^_|_$/g, '')
            .toLowerCase()
            .substring(0, 50) || 'conversation';
    }

    /**
     * Escape HTML characters
     */
    escapeHtml(text) {
        // Manual HTML escaping for web worker context (no DOM available)
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    /**
     * Process file attachment
     */
    async processFileAttachment(file) {
        try {
            const result = {
                name: file.name,
                size: file.size,
                type: file.type,
                lastModified: file.lastModified
            };

            // Determine attachment type
            if (file.type.startsWith('image/')) {
                result.attachmentType = 'image';
                result.dataUrl = await this.fileToDataUrl(file);
            } else if (file.type === 'application/pdf') {
                result.attachmentType = 'pdf';
                result.dataUrl = await this.fileToDataUrl(file);
            } else if (file.type.startsWith('audio/')) {
                result.attachmentType = 'audio';
                result.dataUrl = await this.fileToDataUrl(file);
            } else {
                result.attachmentType = 'file';
                result.dataUrl = await this.fileToDataUrl(file);
            }

            return {
                success: true,
                data: result
            };
        } catch (error) {
            return {
                success: false,
                error: error.message
            };
        }
    }

    /**
     * Convert file to data URL
     */
    fileToDataUrl(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = e => resolve(e.target.result);
            reader.onerror = e => reject(new Error('Failed to read file'));
            reader.readAsDataURL(file);
        });
    }
}

// Worker event handling
const messageProcessor = new MessageProcessor();

self.onmessage = async function(e) {
    const { type, data, id } = e.data;

    try {
        let result;

        switch (type) {
            case 'processMessageContent':
                result = messageProcessor.processMessageContent(data);
                break;

            case 'processMessagesForSearch':
                result = messageProcessor.processMessagesForSearch(data.messages, data.searchTerm);
                break;

            case 'processConversationForExport':
                result = messageProcessor.processConversationForExport(data.conversation, data.format);
                break;

            case 'processFileAttachment':
                result = await messageProcessor.processFileAttachment(data);
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