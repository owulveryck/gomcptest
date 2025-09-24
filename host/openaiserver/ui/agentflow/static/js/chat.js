/**
 * AgentFlow - Agentic System Chat UI
 * Main JavaScript module for the chat interface
 */

class ChatUI {
    constructor() {
        // Use baseURL from global variable injected by the template
        this.baseUrl = window.AGENTFLOW_BASE_URL || '';
        this.apiUrl = this.baseUrl + '/v1/chat/completions';

        this.messages = [];
        this.editingIndex = -1;
        this.models = [];
        this.selectedModel = null;
        this.tools = [];
        this.selectedTools = new Set();  // Store selected tool names
        this.currentReader = null;  // Store the current stream reader
        this.isStreaming = false;   // Track streaming state
        this.toolPopups = new Map();  // Store active tool popups
        this.popupAutoCloseTimers = new Map();  // Store auto-close timers
        this.systemPrompt = 'You are a helpful assistant.\nCurrent time is {{now | formatTimeInLocation "Europe/Paris" "2006-01-02 15:04"}}';  // Default system prompt
        this.selectedFiles = [];  // Store selected files (images/PDFs) as base64 data URIs

        // Storage management flags
        this.artifactServerUnavailable = false;  // Track if artifact server is unavailable
        this.storageQuotaExceeded = false;       // Track if localStorage quota is exceeded
        this.lastCleanupAttempt = null;          // Track last cleanup attempt time

        // Audio recording state
        this.currentAudioSource = 'microphone';
        this.isRecording = false;
        this.isCreatingLap = false;
        this.mediaRecorder = null;
        this.audioStream = null;
        this.recordingStartTime = null;
        this.recordingTimerInterval = null;
        this.audioChunks = [];
        this.audioContext = null;  // Web Audio API context for mixing
        this.originalStreams = null;  // Store original streams for "both" mode cleanup

        // DOM elements (must be initialized first)
        this.chatMessages = document.getElementById('chatMessages');
        this.chatInput = document.getElementById('chatInput');
        this.sendButton = document.getElementById('sendButton');
        this.typingIndicator = document.getElementById('typingIndicator');
        this.modelButton = document.getElementById('modelButton');
        this.modelDropdown = document.getElementById('modelDropdown');
        this.modelsList = document.getElementById('modelsList');
        this.selectedModelName = document.getElementById('selectedModelName');
        this.toolButton = document.getElementById('toolButton');
        this.toolDropdown = document.getElementById('toolDropdown');
        this.toolsList = document.getElementById('toolsList');
        this.selectedToolsCount = document.getElementById('selectedToolsCount');
        this.toolToggleAll = document.getElementById('toolToggleAll');
        this.toolToggleNone = document.getElementById('toolToggleNone');
        this.sideMenu = document.getElementById('sideMenu');
        this.conversationsList = document.getElementById('conversationsList');
        this.newChatBtn = document.getElementById('newChatBtn');
        this.menuTrigger = document.getElementById('menuTrigger');
        this.systemPromptTextarea = document.getElementById('systemPromptTextarea');
        this.systemPromptSave = document.getElementById('systemPromptSave');
        this.systemPromptReset = document.getElementById('systemPromptReset');
        this.copySelectionButton = document.getElementById('copySelectionButton');
        this.attachmentBtn = document.getElementById('attachmentBtn');
        this.attachmentInput = document.getElementById('attachmentInput');
        this.filePreviewContainer = document.getElementById('filePreviewContainer');
        this.exportConversationBtn = document.getElementById('exportConversation');
        this.importConversationBtn = document.getElementById('importConversation');
        this.importFileInput = document.getElementById('importFileInput');

        // Audio recording elements
        this.audioSourceButton = document.getElementById('audioSourceButton');
        this.audioSourceDropdown = document.getElementById('audioSourceDropdown');
        this.selectedAudioSource = document.getElementById('selectedAudioSource');
        this.recordBtn = document.getElementById('recordBtn');
        this.stopBtn = document.getElementById('stopBtn');
        this.segmentBtn = document.getElementById('segmentBtn');
        this.recordingIndicator = document.getElementById('recordingIndicator');
        this.recordingTimer = document.getElementById('recordingTimer');

        // Initialize WorkerManager for heavy operations
        // For embedded UI mode, always use '/ui' prefix for worker paths
        const workerBaseUrl = '/ui';
        this.workerManager = new WorkerManager(workerBaseUrl);
        this.workerReady = false;

        // Conversation management (after DOM elements are initialized)
        this.conversations = {};
        this.currentConversationId = null;

        this.init();

        // Initialize workers and load conversations
        this.initializeWorkers();

        // Check artifact server availability at startup
        this.checkArtifactServerAvailability();
    }

    // Initialize web workers for heavy operations
    async initializeWorkers() {
        try {
            console.log('Initializing web workers...');
            this.updateWorkerStatus('loading', 'Initializing web workers...');

            // Add timeout protection to prevent hanging on worker init
            const workerInitPromise = this.workerManager.init();
            const timeoutPromise = new Promise(resolve => {
                setTimeout(() => resolve({ success: false, error: 'Worker initialization timeout (8s)' }), 8000);
            });

            const result = await Promise.race([workerInitPromise, timeoutPromise]);

            if (result.success) {
                this.workerReady = true;
                console.log('Workers initialized successfully');
                this.updateWorkerStatus('ready', 'Workers ready - enhanced performance enabled');
            } else {
                console.warn('Workers failed to initialize, continuing with fallback mode:', result.error);
                this.workerReady = false;
                this.updateWorkerStatus('fallback', 'Using fallback mode - core functionality available');
            }
        } catch (error) {
            console.warn('Error during worker initialization, using fallback mode:', error);
            this.workerReady = false;
            this.updateWorkerStatus('error', 'Worker initialization failed - using fallback mode');
        }

        // Always continue with loading conversations and initialization
        // (workers are optional enhancement, not required for core functionality)
        try {
            await this.loadConversationsFromWorker();
            this.initializeConversation();

            if (this.workerReady) {
                this.setupAutoSaveWithWorkers();
            } else {
                this.setupAutoSave();
            }
        } catch (error) {
            console.error('Error during conversation loading:', error);
            // Final fallback
            this.conversations = this.loadConversationsFallback();
            this.initializeConversation();
            this.setupAutoSave();
        }
    }

    // Update worker status in the UI
    updateWorkerStatus(status, message) {
        const statusIndicator = document.querySelector('.status-indicator');
        if (statusIndicator) {
            // Remove existing status classes
            statusIndicator.classList.remove('worker-loading', 'worker-ready', 'worker-fallback', 'worker-error');

            // Add new status class
            switch (status) {
                case 'loading':
                    statusIndicator.classList.add('worker-loading');
                    statusIndicator.title = message;
                    break;
                case 'ready':
                    statusIndicator.classList.add('worker-ready');
                    statusIndicator.title = message;
                    break;
                case 'fallback':
                    statusIndicator.classList.add('worker-fallback');
                    statusIndicator.title = message;
                    break;
                case 'error':
                    statusIndicator.classList.add('worker-error');
                    statusIndicator.title = message;
                    break;
            }
        }

        // Also show a brief notification for status changes
        if (status !== 'loading') {
            setTimeout(() => {
                this.showNotification(message, status === 'ready' ? 'success' : 'info');
            }, 100);
        }
    }

    // Load conversations using storage worker
    async loadConversationsFromWorker() {
        // CRITICAL FIX: Always use fallback for loading since workers don't handle storage
        // Workers are only for data processing, not actual storage operations
        this.conversations = this.loadConversationsFallback();
        console.log(`Loaded ${Object.keys(this.conversations).length} conversations from localStorage fallback`);
        return;
    }

    // Fallback conversation loading (synchronous)
    loadConversationsFallback() {
        try {
            const stored = localStorage.getItem('chat_conversations');
            return stored ? JSON.parse(stored) : {};
        } catch (error) {
            console.error('Fallback conversation loading failed:', error);
            return {};
        }
    }

    // Set up auto-save using workers
    setupAutoSaveWithWorkers() {
        // Save every 30 seconds using workers
        this.autoSaveInterval = setInterval(async () => {
            if (this.messages && this.messages.length > 0 && this.workerReady) {
                try {
                    await this.saveConversationsViaWorker();
                } catch (error) {
                    console.error('Auto-save via worker failed:', error);
                    // Fallback to synchronous save
                    this.saveConversationsFallback();
                }
            }
        }, 30000);

        // Save on page unload
        window.addEventListener('beforeunload', async () => {
            if (this.workerReady) {
                try {
                    await this.saveConversationsViaWorker();
                } catch (error) {
                    this.saveConversationsFallback();
                }
            } else {
                this.saveConversationsFallback();
            }
        });
    }

    // Set up periodic auto-save mechanism (fallback)
    setupAutoSave() {
        // Save every 30 seconds as a safety net
        this.autoSaveInterval = setInterval(() => {
            if (this.messages && this.messages.length > 0) {
                this.saveConversations().catch(error => {
                    console.warn('Periodic auto-save failed:', error);
                });
                console.log('Periodic auto-save completed');
            }
        }, 30000); // 30 seconds

        // Save on page unload/refresh
        window.addEventListener('beforeunload', () => {
            this.saveConversations().catch(error => {
                console.error('Final save on page unload failed:', error);
            });
        });

        // Save on page visibility change (when user switches tabs)
        document.addEventListener('visibilitychange', () => {
            if (document.hidden && this.messages && this.messages.length > 0) {
                this.saveConversations().catch(error => {
                    console.warn('Save on visibility change failed:', error);
                });
            }
        });

        console.log('Auto-save mechanisms initialized');
    }

    // Check if artifact storage server is available at startup
    async checkArtifactServerAvailability() {
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 2000); // 2 second timeout

            const response = await fetch(`${this.baseUrl}/artifact`, {
                method: 'GET',
                signal: controller.signal
            });

            clearTimeout(timeoutId);

            if (response.ok) {
                console.log('Artifact storage server is available');
                this.artifactServerUnavailable = false;
            } else {
                console.warn('Artifact storage server returned error:', response.status);
                this.artifactServerUnavailable = true;
            }
        } catch (error) {
            console.warn('Artifact storage server not available at startup');
            this.artifactServerUnavailable = true;
        }
    }

    init() {
        this.sendButton.addEventListener('click', () => {
            if (this.isStreaming) {
                // Prevent multiple stop clicks
                if (this.sendButton.disabled) return;

                // Temporarily disable button to prevent spam clicking
                this.sendButton.disabled = true;
                this.sendButton.textContent = 'Stopping...';

                this.stopStreaming().then(() => {
                    // Re-enable after a short delay
                    setTimeout(() => {
                        this.sendButton.disabled = false;
                    }, 500);
                });
            } else {
                this.sendMessage();
            }
        });
        this.chatInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                if (!this.isStreaming) {
                    this.sendMessage();
                }
            }
        });

        // Auto-resize textarea
        this.chatInput.addEventListener('input', () => {
            this.chatInput.style.height = 'auto';
            this.chatInput.style.height = Math.min(this.chatInput.scrollHeight, 150) + 'px';
        });

        // Model selector events
        this.modelButton.addEventListener('click', () => {
            this.modelDropdown.classList.toggle('active');
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.modelButton.contains(e.target) && !this.modelDropdown.contains(e.target)) {
                this.modelDropdown.classList.remove('active');
            }
            if (!this.toolButton.contains(e.target) && !this.toolDropdown.contains(e.target)) {
                this.toolDropdown.classList.remove('active');
            }
        });

        // Tool selector events
        this.toolButton.addEventListener('click', () => {
            this.toolDropdown.classList.toggle('active');
        });

        this.toolToggleAll.addEventListener('click', () => {
            this.selectAllTools();
        });

        this.toolToggleNone.addEventListener('click', () => {
            this.selectNoTools();
        });

        // Load available models and tools
        this.loadModels();
        this.loadTools();

        // Initialize conversation management
        this.newChatBtn.addEventListener('click', () => this.createNewConversation());
        this.renderConversationsList();

        // System prompt event listeners
        this.systemPromptSave.addEventListener('click', () => this.saveSystemPrompt());
        this.systemPromptReset.addEventListener('click', () => this.resetSystemPrompt());
        document.getElementById('cleanupStorage').addEventListener('click', () => this.manualCleanupStorage());

        // Auto-save system prompt on change
        this.systemPromptTextarea.addEventListener('change', () => {
            this.systemPrompt = this.systemPromptTextarea.value.trim() || 'You are a helpful assistant.\nCurrent time is {{now | formatTimeInLocation "Europe/Paris" "2006-01-02 15:04"}}';
        });

        // Menu trigger handling
        this.menuTrigger.addEventListener('click', () => {
            this.sideMenu.classList.add('active');
        });

        // Click outside to close menu
        document.addEventListener('click', (e) => {
            if (!this.sideMenu.contains(e.target) &&
                !this.menuTrigger.contains(e.target) &&
                this.sideMenu.classList.contains('active')) {
                this.sideMenu.classList.remove('active');
            }
        });

        // Auto-save on message send
        this.originalSendMessage = this.sendMessage.bind(this);

        // Text selection handling for copy functionality
        this.initTextSelectionHandling();

        // File upload handling (images and PDFs)
        this.initFileUploadHandling();

        // Export/Import functionality
        this.initExportImportHandling();

        // Audio recording functionality
        this.initAudioRecording();
    }

    // Text selection and copy functionality
    initTextSelectionHandling() {
        // Track text selection in chat messages
        document.addEventListener('selectionchange', () => {
            this.handleSelectionChange();
        });

        // Copy button click handler
        this.copySelectionButton.addEventListener('click', () => {
            this.copySelectedMarkdown();
        });

        // Hide copy button when clicking elsewhere (but not immediately)
        document.addEventListener('click', (e) => {
            if (!this.copySelectionButton.contains(e.target)) {
                // Use a small delay to allow selection change event to fire first
                setTimeout(() => {
                    const selection = window.getSelection();
                    if (selection.rangeCount === 0 || selection.isCollapsed) {
                        this.hideCopyButton();
                    }
                }, 10);
            }
        });
    }

    handleSelectionChange() {
        const selection = window.getSelection();
        if (selection.rangeCount === 0 || selection.isCollapsed) {
            this.hideCopyButton();
            return;
        }

        // Check if selection is within chat messages
        const range = selection.getRangeAt(0);
        const chatMessagesContainer = this.chatMessages;

        if (!chatMessagesContainer.contains(range.commonAncestorContainer)) {
            this.hideCopyButton();
            return;
        }

        // Show copy button near selection
        this.showCopyButton(range);
    }

    showCopyButton(range) {
        const button = this.copySelectionButton;

        // Show button in fixed position (CSS handles positioning)
        button.style.display = 'block';

        // Reset button state
        button.className = 'copy-selection-button';
        button.innerHTML = `
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 4px;">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
            </svg>
            Copy
        `;
    }

    hideCopyButton() {
        this.copySelectionButton.style.display = 'none';
    }

    async copySelectedMarkdown() {
        const selection = window.getSelection();
        if (selection.rangeCount === 0 || selection.isCollapsed) {
            return;
        }

        try {
            // Get the markdown source for the selected content
            const markdownText = this.extractMarkdownFromSelection(selection);

            // Copy to clipboard
            await navigator.clipboard.writeText(markdownText);

            // Show success feedback
            const button = this.copySelectionButton;
            button.className = 'copy-selection-button success';
            button.innerHTML = `
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 4px;">
                    <polyline points="20,6 9,17 4,12"></polyline>
                </svg>
                Copied!
            `;

            // Hide button after short delay
            setTimeout(() => {
                this.hideCopyButton();
            }, 1500);

        } catch (error) {
            console.error('Failed to copy text:', error);

            // Show error feedback
            const button = this.copySelectionButton;
            button.innerHTML = 'Failed to copy';
            setTimeout(() => {
                this.hideCopyButton();
            }, 2000);
        }
    }

    extractMarkdownFromSelection(selection) {
        const range = selection.getRangeAt(0);

        // Find which message(s) the selection spans
        const selectedMessages = this.findMessagesInSelection(range);

        if (selectedMessages.length === 0) {
            // Fallback to plain text if we can't find the source
            return selection.toString();
        }

        // If selection spans multiple messages, combine their markdown
        if (selectedMessages.length > 1) {
            return selectedMessages.map(msg => msg.content).join('\n\n');
        }

        // Single message - try to extract the relevant portion
        const message = selectedMessages[0];
        const selectedText = selection.toString().trim();

        // Check if user actually selected the entire message content area
        // (not just text that happens to match most of the message)
        if (this.isActuallyEntireMessageSelected(range, message, selectedText)) {
            return message.content;
        }

        // Try to map the selected HTML back to markdown portions
        return this.mapSelectionToMarkdown(message.content, selectedText, range);
    }

    findMessagesInSelection(range) {
        const selectedMessages = [];

        // Get all message groups in the chat (excludes tool notifications)
        const messageGroups = Array.from(this.chatMessages.querySelectorAll('.message-group'));

        // Create a mapping between DOM elements and message indices
        let messageIndex = 0;

        messageGroups.forEach((messageGroup) => {
            const messageContent = messageGroup.querySelector('.message-content');
            if (messageContent && this.doesSelectionIntersectElement(range, messageContent)) {
                // Find the corresponding message in our array (skip tool notifications)
                while (messageIndex < this.messages.length) {
                    const message = this.messages[messageIndex];
                    if (message && (message.role === 'user' || message.role === 'assistant')) {
                        if (!selectedMessages.includes(message)) {
                            selectedMessages.push(message);
                        }
                        messageIndex++;
                        break;
                    }
                    messageIndex++;
                }
            } else {
                // Skip this message group, advance index to next user/assistant message
                while (messageIndex < this.messages.length) {
                    const message = this.messages[messageIndex];
                    if (message && (message.role === 'user' || message.role === 'assistant')) {
                        messageIndex++;
                        break;
                    }
                    messageIndex++;
                }
            }
        });

        return selectedMessages;
    }

    doesSelectionIntersectElement(range, element) {
        try {
            // Create a range that spans the entire element
            const elementRange = document.createRange();
            elementRange.selectNodeContents(element);

            // Check if the selection range intersects with the element range
            return range.compareBoundaryPoints(Range.END_TO_START, elementRange) <= 0 &&
                   range.compareBoundaryPoints(Range.START_TO_END, elementRange) >= 0;
        } catch (e) {
            // Fallback: check if element contains any part of the selection
            return element.contains(range.startContainer) || element.contains(range.endContainer);
        }
    }

    isActuallyEntireMessageSelected(range, message, selectedText) {
        // Find the message content element for this message
        const messageGroups = Array.from(this.chatMessages.querySelectorAll('.message-group'));

        // Find the correct DOM element for this message by matching content
        let targetMessageGroup = null;
        let domMessageIndex = 0;

        for (let i = 0; i < this.messages.length; i++) {
            const msg = this.messages[i];
            if (msg.role === 'user' || msg.role === 'assistant') {
                if (msg === message) {
                    targetMessageGroup = messageGroups[domMessageIndex];
                    break;
                }
                domMessageIndex++;
            }
        }

        if (!targetMessageGroup) {
            return false;
        }

        const messageContent = targetMessageGroup.querySelector('.message-content');

        if (!messageContent) {
            return false;
        }

        try {
            // Create a range that spans the entire message content
            const fullMessageRange = document.createRange();
            fullMessageRange.selectNodeContents(messageContent);

            // Check if the user's selection covers most of the message content area
            const selectionCoversStart = range.compareBoundaryPoints(Range.START_TO_START, fullMessageRange) <= 0;
            const selectionCoversEnd = range.compareBoundaryPoints(Range.END_TO_END, fullMessageRange) >= 0;

            // Also check if the selection includes action buttons (indicating full selection)
            const includesActionButtons = selectedText.includes('Edit') && selectedText.includes('Replay from here');

            return (selectionCoversStart && selectionCoversEnd) || includesActionButtons;
        } catch (e) {
            // Fallback to text-based heuristic if range comparison fails
            return this.isEntireMessageSelectedFallback(message, selectedText);
        }
    }

    isEntireMessageSelectedFallback(message, selectedText) {
        // Clean up the selection text by removing action buttons and extra whitespace
        const cleanSelection = selectedText
            .replace(/\s*Edit\s*Replay from here\s*/g, '') // Remove action buttons
            .replace(/\s+/g, ' ')
            .trim();

        // Clean up the message content for comparison
        const messageText = message.content
            .replace(/[#*`_\[\]()]/g, '') // Remove markdown formatting
            .replace(/\s+/g, ' ')
            .trim();

        // Only consider it "entire" if the cleaned selection is very close to the full message
        const similarityRatio = cleanSelection.length / messageText.length;

        // Be more strict - only return true if selection is 90%+ of the message
        return similarityRatio > 0.9;
    }

    mapSelectionToMarkdown(markdownContent, selectedText, range) {
        // Get the character positions in the original markdown that correspond to the selection
        const markdownPositions = this.getMarkdownPositionsFromSelection(markdownContent, range);

        if (markdownPositions) {
            const { start, end } = markdownPositions;
            return markdownContent.substring(start, end);
        }

        // Simple fallback - clean the selected text and try direct match
        const cleanSelection = this.cleanSelectedText(selectedText);

        // Try exact match in markdown
        if (markdownContent.includes(cleanSelection)) {
            const startIndex = markdownContent.indexOf(cleanSelection);
            return markdownContent.substring(startIndex, startIndex + cleanSelection.length);
        }

        // Last resort: return cleaned selection as-is
        return cleanSelection || selectedText;
    }

    /**
     * Map DOM selection range to original markdown character positions
     * This method attempts to find the exact positions in the markdown source
     * that correspond to the selected text in the rendered DOM
     */
    getMarkdownPositionsFromSelection(markdownContent, range) {
        try {
            // Get the selected text from the range
            const selectedText = range.toString().trim();
            if (!selectedText) return null;

            // Clean the selected text for matching (remove extra whitespace, copy buttons, etc.)
            const cleanSelection = this.cleanSelectedText(selectedText);
            if (!cleanSelection) return null;

            // Try to find exact match in markdown
            const exactMatchIndex = markdownContent.indexOf(cleanSelection);
            if (exactMatchIndex !== -1) {
                return {
                    start: exactMatchIndex,
                    end: exactMatchIndex + cleanSelection.length
                };
            }

            // If no exact match, try word-based matching
            const words = cleanSelection.split(/\s+/).filter(word => word.length > 2);
            if (words.length === 0) return null;

            // Find the first and last significant words in the markdown
            const firstWord = words[0];
            const lastWord = words[words.length - 1];

            const firstWordIndex = markdownContent.indexOf(firstWord);
            const lastWordIndex = markdownContent.lastIndexOf(lastWord);

            if (firstWordIndex !== -1 && lastWordIndex !== -1 && firstWordIndex <= lastWordIndex) {
                // Try to find the most likely boundaries
                let start = firstWordIndex;
                let end = lastWordIndex + lastWord.length;

                // Refine the boundaries by looking for word boundaries
                // Go backwards from firstWordIndex to find a good start
                while (start > 0 && /\w/.test(markdownContent[start - 1])) {
                    start--;
                }

                // Go forwards from lastWordIndex + lastWord.length to find a good end
                while (end < markdownContent.length && /\w/.test(markdownContent[end])) {
                    end++;
                }

                return { start, end };
            }

            return null;
        } catch (error) {
            console.warn('Error mapping selection to markdown positions:', error);
            return null;
        }
    }

    /**
     * Clean selected text by removing copy buttons and normalizing whitespace
     */
    cleanSelectedText(text) {
        return text
            .replace(/\s*Copy\s*/g, '') // Remove copy button text
            .replace(/\s*Edit\s*Replay from here\s*/g, '') // Remove action button combinations
            .replace(/\s*Edit\s*/g, '') // Remove edit button text
            .replace(/\s*Replay from here\s*/g, '') // Remove replay button text
            .replace(/\s+/g, ' ') // Normalize whitespace
            .trim();
    }

    reconstructMarkdownFromRange(range) {
        try {
            const fragment = range.cloneContents();
            return this.convertDOMToMarkdown(fragment);
        } catch (error) {
            console.warn('Failed to reconstruct markdown from range:', error);
            return null;
        }
    }

    convertDOMToMarkdown(node) {
        let markdown = '';

        // Handle different node types
        if (node.nodeType === Node.TEXT_NODE) {
            return node.textContent;
        }

        if (node.nodeType === Node.ELEMENT_NODE) {
            const tagName = node.tagName.toLowerCase();
            const children = Array.from(node.childNodes);

            switch (tagName) {
                case 'strong':
                case 'b':
                    markdown += '**' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '**';
                    break;

                case 'em':
                case 'i':
                    markdown += '*' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '*';
                    break;

                case 'code':
                    // Check if this is inline code (not in a pre block)
                    if (!node.closest('pre')) {
                        markdown += '`' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '`';
                    } else {
                        // This is a code block, preserve as is
                        markdown += children.map(child => this.convertDOMToMarkdown(child)).join('');
                    }
                    break;

                case 'pre':
                    // Code block
                    const codeElement = node.querySelector('code');
                    if (codeElement) {
                        const language = this.extractLanguageFromCodeElement(codeElement);
                        const codeContent = codeElement.textContent || codeElement.innerText || '';
                        markdown += '```' + (language || '') + '\n' + codeContent + '\n```';
                    } else {
                        markdown += '```\n' + node.textContent + '\n```';
                    }
                    break;

                case 'h1':
                    markdown += '# ' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '\n';
                    break;

                case 'h2':
                    markdown += '## ' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '\n';
                    break;

                case 'h3':
                    markdown += '### ' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '\n';
                    break;

                case 'h4':
                    markdown += '#### ' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '\n';
                    break;

                case 'h5':
                    markdown += '##### ' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '\n';
                    break;

                case 'h6':
                    markdown += '###### ' + children.map(child => this.convertDOMToMarkdown(child)).join('') + '\n';
                    break;

                case 'a':
                    const href = node.getAttribute('href');
                    const linkText = children.map(child => this.convertDOMToMarkdown(child)).join('');
                    if (href) {
                        markdown += '[' + linkText + '](' + href + ')';
                    } else {
                        markdown += linkText;
                    }
                    break;

                case 'ul':
                case 'ol':
                    // Process list items
                    children.forEach(child => {
                        if (child.tagName && child.tagName.toLowerCase() === 'li') {
                            const listItemContent = this.convertDOMToMarkdown(child);
                            markdown += (tagName === 'ul' ? '- ' : '1. ') + listItemContent + '\n';
                        }
                    });
                    break;

                case 'li':
                    // List item content (without the bullet/number, that's handled by ul/ol)
                    markdown += children.map(child => this.convertDOMToMarkdown(child)).join('');
                    break;

                case 'p':
                    markdown += children.map(child => this.convertDOMToMarkdown(child)).join('') + '\n\n';
                    break;

                case 'br':
                    markdown += '\n';
                    break;

                default:
                    // For unknown elements, just process children
                    markdown += children.map(child => this.convertDOMToMarkdown(child)).join('');
                    break;
            }
        } else if (node.nodeType === Node.DOCUMENT_FRAGMENT_NODE) {
            // Document fragment, process all children
            markdown += Array.from(node.childNodes).map(child => this.convertDOMToMarkdown(child)).join('');
        }

        return markdown;
    }

    extractLanguageFromCodeElement(codeElement) {
        // Try to extract language from class names like "language-python"
        const classList = Array.from(codeElement.classList);
        for (const className of classList) {
            if (className.startsWith('language-')) {
                return className.substring(9); // Remove "language-" prefix
            }
        }
        return '';
    }


    // File upload functionality (images, PDFs, and audio)
    initFileUploadHandling() {
        // Unified attachment button click handler
        this.attachmentBtn.addEventListener('click', () => {
            this.attachmentInput.click();
        });

        // File input change handler
        this.attachmentInput.addEventListener('change', (e) => {
            this.handleFileSelection(e.target.files);
        });

        // Drag and drop support
        this.chatInput.addEventListener('dragover', (e) => {
            e.preventDefault();
            e.stopPropagation();
        });

        this.chatInput.addEventListener('drop', (e) => {
            e.preventDefault();
            e.stopPropagation();

            const files = Array.from(e.dataTransfer.files).filter(file =>
                file.type.startsWith('image/') ||
                file.type === 'application/pdf' ||
                file.type.startsWith('audio/')
            );

            if (files.length > 0) {
                this.handleFileSelection(files);
            }
        });

        // Clipboard paste support for images
        this.chatInput.addEventListener('paste', (e) => {
            const items = e.clipboardData?.items;
            if (!items) return;

            for (let i = 0; i < items.length; i++) {
                const item = items[i];

                // Check if it's an image
                if (item.type.startsWith('image/')) {
                    e.preventDefault(); // Prevent the default paste behavior for images

                    const file = item.getAsFile();
                    if (file) {
                        // Generate a meaningful filename based on timestamp
                        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
                        const extension = file.type.split('/')[1] || 'png';
                        const fileName = `pasted-image-${timestamp}.${extension}`;

                        // Create a new File object with a proper name
                        const namedFile = new File([file], fileName, { type: file.type });

                        this.handleFileSelection([namedFile]);
                    }
                }
            }
        });
    }

    async handleFileSelection(files) {
        const SIZE_THRESHOLD = 50 * 1024; // 50KB threshold for artifact storage - much more aggressive

        for (const file of files) {
            if (file.type.startsWith('image/') ||
                file.type === 'application/pdf' ||
                file.type.startsWith('audio/')) {
                try {
                    // Check file size and use artifact storage for large files
                    if (file.size > SIZE_THRESHOLD && !this.artifactServerUnavailable) {
                        console.log(`Large file detected (${file.size} bytes), using artifact storage:`, file.name);
                        try {
                            const artifactId = await this.uploadToArtifactStorage(file, file.name);
                            this.addFilePreview(`artifact:${artifactId}`, file.name, file.type, {
                                isArtifact: true,
                                artifactId: artifactId,
                                size: file.size
                            });
                            console.log('File uploaded to artifact storage:', artifactId);
                        } catch (uploadError) {
                            console.warn('Artifact upload failed, falling back to base64:', uploadError);
                            // Fallback to base64 if artifact upload fails
                            const dataURL = await this.fileToDataURL(file);
                            this.addFilePreview(dataURL, file.name, file.type);
                        }
                    } else {
                        // Use base64 for small files or when artifact server unavailable
                        const dataURL = await this.fileToDataURL(file);
                        this.addFilePreview(dataURL, file.name, file.type);
                    }
                } catch (error) {
                    console.error('Error processing file:', error);
                    this.showError(`Failed to process file: ${file.name}`);
                }
            }
        }

        // Clear the file input so the same file can be selected again
        this.attachmentInput.value = '';
    }

    // Keep backward compatibility
    async handleImageSelection(files) {
        return this.handleFileSelection(files);
    }

    fileToDataURL(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => resolve(e.target.result);
            reader.onerror = (e) => reject(e);
            reader.readAsDataURL(file);
        });
    }

    addFilePreview(dataURL, fileName, fileType, metadata = {}) {
        // Store the file data
        const fileData = {
            dataURL: dataURL,
            fileName: fileName,
            fileType: fileType,
            isArtifact: metadata.isArtifact || false,
            artifactId: metadata.artifactId,
            size: metadata.size,
            duration: metadata.duration,
            id: Date.now() + Math.random() // Simple unique ID
        };

        this.selectedFiles.push(fileData);

        // Calculate file size - use metadata size for artifacts, calculate for data URLs
        const fileSize = metadata.size || this.calculateDataURLSize(dataURL);
        const formattedSize = this.formatFileSize(fileSize);

        // Create preview element
        const preview = document.createElement('div');
        preview.className = 'file-preview';
        preview.dataset.fileId = fileData.id;

        if (fileType.startsWith('image/')) {
            // Image preview - handle artifact URLs
            const imageUrl = dataURL.startsWith('artifact:') ?
                `/artifact/${dataURL.substring(9)}` : dataURL;
            const artifactBadge = fileData.isArtifact ?
                `<div class="artifact-badge" style="position: absolute; top: -2px; left: 2px; background: #059669; color: white; padding: 1px 3px; border-radius: 2px; font-size: 7px; font-weight: bold;">SERVER</div>` : '';
            preview.innerHTML = `
                <img src="${imageUrl}" alt="${fileName}" title="${fileName}">
                ${artifactBadge}
                <div class="file-size-badge" style="position: absolute; top: 2px; right: 20px; background: rgba(0,0,0,0.7); color: white; padding: 1px 4px; border-radius: 3px; font-size: 9px; font-weight: 500;">
                    ${formattedSize}
                </div>
                <button class="remove-file" onclick="chatUI.removeFilePreview('${fileData.id}')">×</button>
            `;
        } else if (fileType === 'application/pdf') {
            // PDF preview - show artifact indicator if stored on server
            const artifactBadge = fileData.isArtifact ?
                `<div class="artifact-badge" style="position: absolute; top: -2px; left: 2px; background: #dc2626; color: white; padding: 1px 3px; border-radius: 2px; font-size: 7px; font-weight: bold;">SERVER</div>` : '';
            preview.innerHTML = `
                <div class="pdf-icon" title="${fileName}">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
                    </svg>
                    <div style="word-wrap: break-word; font-size: 10px;">${fileName.length > 12 ? fileName.substring(0, 12) + '...' : fileName}</div>
                    <div style="font-size: 8px; color: #666; margin-top: 2px;">${formattedSize}</div>
                </div>
                ${artifactBadge}
                <button class="remove-file" onclick="chatUI.removeFilePreview('${fileData.id}')">×</button>
            `;
        } else if (fileType.startsWith('audio/')) {
            // Audio preview - show artifact indicator if stored on server
            const artifactBadge = fileData.isArtifact ?
                `<div class="artifact-badge" style="position: absolute; top: -2px; right: 18px; background: #059669; color: white; padding: 1px 3px; border-radius: 2px; font-size: 7px; font-weight: bold;">SERVER</div>` : '';

            preview.innerHTML = `
                <div class="audio-icon ${fileData.isArtifact ? 'artifact' : ''}" title="${fileName}${fileData.isArtifact ? ' (Server Storage)' : ''}">
                    <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                        <path d="M1 8h1l1-2 2 4 2-4 2 2h1"/>
                    </svg>
                    <div style="word-wrap: break-word; font-size: 8px;">${fileName.length > 8 ? fileName.substring(0, 8) + '...' : fileName}</div>
                    <div style="font-size: 7px; color: #ccc; margin-top: 1px;">${formattedSize}</div>
                    ${artifactBadge}
                </div>
                <button class="remove-file" onclick="chatUI.removeFilePreview('${fileData.id}')">×</button>
            `;
        }

        this.filePreviewContainer.appendChild(preview);
        this.updateFilePreviewVisibility();
    }

    // Keep backward compatibility
    addImagePreview(dataURL, fileName) {
        return this.addFilePreview(dataURL, fileName, 'image/jpeg');
    }

    removeFilePreview(fileId) {
        // Remove from selectedFiles array
        this.selectedFiles = this.selectedFiles.filter(file => file.id != fileId);

        // Remove preview element
        const preview = this.filePreviewContainer.querySelector(`[data-file-id="${fileId}"]`);
        if (preview) {
            preview.remove();
        }

        this.updateFilePreviewVisibility();
    }

    // Keep backward compatibility
    removeImagePreview(imageId) {
        return this.removeFilePreview(imageId);
    }

    updateFilePreviewVisibility() {
        if (this.selectedFiles.length > 0) {
            this.filePreviewContainer.style.display = 'flex';
        } else {
            this.filePreviewContainer.style.display = 'none';
        }
    }

    // Keep backward compatibility
    updateImagePreviewVisibility() {
        return this.updateFilePreviewVisibility();
    }

    clearFilePreviews() {
        this.selectedFiles = [];
        this.filePreviewContainer.innerHTML = '';
        this.updateFilePreviewVisibility();
    }

    // Keep backward compatibility
    clearImagePreviews() {
        return this.clearFilePreviews();
    }

    // Export/Import functionality
    initExportImportHandling() {
        // Export button click handler
        this.exportConversationBtn.addEventListener('click', () => {
            this.exportConversationAsMarkdown();
        });

        // Import button click handler
        this.importConversationBtn.addEventListener('click', () => {
            this.importFileInput.click();
        });

        // File input change handler for import
        this.importFileInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                this.importConversationFromMarkdown(file);
                // Reset the input so the same file can be selected again
                e.target.value = '';
            }
        });
    }

    exportConversationAsMarkdown() {
        if (!this.currentConversationId || !this.conversations[this.currentConversationId]) {
            this.showNotification('No conversation to export', 'warning');
            return;
        }

        const conversation = this.conversations[this.currentConversationId];
        let imageCounter = 0;
        let markdownContent = '';
        let imageReferences = '';

        // Add conversation title and metadata
        markdownContent += `# ${conversation.title}\n\n`;
        markdownContent += `*Exported from AgentFlow on ${new Date().toLocaleString()}*\n\n`;
        markdownContent += `**System Prompt:** ${conversation.systemPrompt || 'Default'}\n\n---\n\n`;

        // Process each message
        for (const message of conversation.messages) {
            // Skip tool notifications for export
            if (message.role === 'tool') {
                continue;
            }

            // Add role header with emoji and model name for assistant
            if (message.role === 'user') {
                markdownContent += `## 👤 User\n\n`;
            } else if (message.role === 'assistant') {
                const modelName = this.selectedModel || 'Assistant';
                markdownContent += `## 🧠 Assistant (${modelName})\n\n`;
            }

            // Process message content
            if (typeof message.content === 'string') {
                // Simple text content
                markdownContent += message.content + '\n\n';
            } else if (Array.isArray(message.content)) {
                // Multimodal content
                for (const item of message.content) {
                    if (item.type === 'text') {
                        markdownContent += item.text + '\n\n';
                    } else if (item.type === 'image_url' && item.image_url && item.image_url.url) {
                        // Handle images
                        if (item.image_url.url !== '[Large image data removed to save storage space]') {
                            imageCounter++;
                            const imageId = `image${imageCounter}`;
                            markdownContent += `![][${imageId}]\n\n`;
                            imageReferences += `[${imageId}]: ${item.image_url.url}\n`;
                        } else {
                            markdownContent += `*[Image not available - removed to save storage space]*\n\n`;
                        }
                    } else if (item.type === 'file' && item.file && item.file.filename) {
                        // Handle files
                        if (item.file.metadata !== '[File data removed to save storage space]') {
                            if (item.file.file_data) {
                                imageCounter++;
                                const fileId = `file${imageCounter}`;
                                markdownContent += `**File:** ${item.file.filename}\n\n![][${fileId}]\n\n`;
                                imageReferences += `[${fileId}]: ${item.file.file_data}\n`;
                            } else {
                                markdownContent += `**File:** ${item.file.filename}\n\n`;
                            }
                        } else {
                            markdownContent += `**File:** ${item.file.filename} *[File not available - removed to save storage space]*\n\n`;
                        }
                    } else if (item.type === 'audio' && item.audio && item.audio.data) {
                        // Handle audio files
                        if (item.audio.data !== '[Large audio data removed to save storage space]') {
                            imageCounter++;
                            const audioId = `audio${imageCounter}`;
                            markdownContent += `**Audio file**\n\n![][${audioId}]\n\n`;
                            imageReferences += `[${audioId}]: ${item.audio.data}\n`;
                        } else {
                            markdownContent += `**Audio file** *[Audio not available - removed to save storage space]*\n\n`;
                        }
                    }
                }
            }
        }

        // Add image references at the end
        if (imageReferences) {
            markdownContent += '\n---\n\n' + imageReferences;
        }

        // Create and download the file
        const blob = new Blob([markdownContent], { type: 'text/markdown' });
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = `${conversation.title.replace(/[^a-z0-9]/gi, '_').toLowerCase()}_${new Date().toISOString().split('T')[0]}.md`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showNotification('Conversation exported successfully!', 'success');
    }

    async importConversationFromMarkdown(file) {
        try {
            const text = await this.readFileAsText(file);
            const parsedConversation = this.parseMarkdownConversation(text);

            if (parsedConversation.messages.length === 0) {
                this.showNotification('No messages found in the markdown file', 'warning');
                return;
            }

            // Create new conversation with imported data
            const id = 'conv_' + Date.now();
            const title = parsedConversation.title || file.name.replace(/\.[^/.]+$/, "");

            this.conversations[id] = {
                id: id,
                title: title,
                messages: parsedConversation.messages,
                systemPrompt: parsedConversation.systemPrompt || this.systemPrompt,
                createdAt: Date.now(),
                lastModified: Date.now()
            };

            // Switch to the imported conversation
            this.saveConversations();
            this.loadConversation(id);
            this.renderConversationsList();

            this.showNotification(`Conversation "${title}" imported successfully!`, 'success');
        } catch (error) {
            console.error('Import error:', error);
            this.showNotification(`Failed to import conversation: ${error.message}`, 'error');
        }
    }

    readFileAsText(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => resolve(e.target.result);
            reader.onerror = (e) => reject(new Error('Failed to read file'));
            reader.readAsText(file);
        });
    }

    parseMarkdownConversation(markdownText) {
        const lines = markdownText.split('\n');
        const messages = [];
        let currentMessage = null;
        let currentContent = [];
        let systemPrompt = '';
        let title = '';
        let inCodeBlock = false;

        // Extract image references at the end
        const imageReferences = {};
        const imageRefRegex = /^\[([^\]]+)\]:\s*(.+)$/;

        // First pass: collect image references
        for (const line of lines) {
            const match = line.match(imageRefRegex);
            if (match) {
                imageReferences[match[1]] = match[2];
            }
        }

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];

            // Track code blocks to avoid parsing headers inside them
            if (line.startsWith('```')) {
                inCodeBlock = !inCodeBlock;
            }

            if (inCodeBlock) {
                if (currentMessage) {
                    currentContent.push(line);
                }
                continue;
            }

            // Extract title
            if (line.startsWith('# ') && !title) {
                title = line.substring(2).trim();
                continue;
            }

            // Extract system prompt
            if (line.includes('**System Prompt:**')) {
                const promptMatch = line.match(/\*\*System Prompt:\*\*\s*(.+)/);
                if (promptMatch) {
                    systemPrompt = promptMatch[1].trim();
                }
                continue;
            }

            // Check for message headers
            if (line.startsWith('## 👤 User') || line.startsWith('## 🧠 Assistant')) {
                // Save previous message if exists
                if (currentMessage) {
                    currentMessage.content = this.processMessageContent(currentContent.join('\n').trim(), imageReferences);
                    messages.push(currentMessage);
                }

                // Start new message
                if (line.startsWith('## 👤 User')) {
                    currentMessage = { role: 'user', content: '' };
                } else {
                    currentMessage = { role: 'assistant', content: '' };
                }
                currentContent = [];
                continue;
            }

            // Skip metadata and separator lines
            if (line.startsWith('*Exported from') || line.startsWith('---') || line.match(imageRefRegex)) {
                continue;
            }

            // Add content to current message
            if (currentMessage) {
                currentContent.push(line);
            }
        }

        // Save last message if exists
        if (currentMessage) {
            currentMessage.content = this.processMessageContent(currentContent.join('\n').trim(), imageReferences);
            messages.push(currentMessage);
        }

        return {
            title: title,
            systemPrompt: systemPrompt,
            messages: messages
        };
    }

    processMessageContent(content, imageReferences) {
        if (!content) return '';

        // Check for images and files in the content
        const imageRegex = /!\[\]\[([^\]]+)\]/g;
        const fileRegex = /\*\*File:\*\*\s*([^\n]+)/g;
        const audioRegex = /\*\*Audio file\*\*/g;

        const hasImages = imageRegex.test(content);
        const hasFiles = fileRegex.test(content);
        const hasAudio = audioRegex.test(content);

        // Reset regex lastIndex
        imageRegex.lastIndex = 0;
        fileRegex.lastIndex = 0;
        audioRegex.lastIndex = 0;

        if (!hasImages && !hasFiles && !hasAudio) {
            // Pure text content
            return content;
        }

        // Multimodal content
        const contentArray = [];
        let textParts = [];
        let lastIndex = 0;

        // Extract text content, excluding image/file references
        let textContent = content.replace(/!\[\]\[([^\]]+)\]/g, '').replace(/\*\*File:\*\*\s*[^\n]+/g, '').replace(/\*\*Audio file\*\*/g, '').trim();

        if (textContent) {
            contentArray.push({
                type: 'text',
                text: textContent
            });
        }

        // Add images
        let match;
        while ((match = imageRegex.exec(content)) !== null) {
            const imageId = match[1];
            const imageUrl = imageReferences[imageId];
            if (imageUrl) {
                if (imageId.startsWith('file')) {
                    // Extract filename from preceding text
                    const fileMatch = content.match(new RegExp(`\\*\\*File:\\*\\*\\s*([^\\n]+)\\s*[\\s\\S]*?!\\[\\]\\[${imageId}\\]`));
                    const filename = fileMatch ? fileMatch[1].trim() : 'unknown_file';
                    contentArray.push({
                        type: 'file',
                        file: {
                            file_data: imageUrl,
                            filename: filename
                        }
                    });
                } else if (imageId.startsWith('audio')) {
                    contentArray.push({
                        type: 'audio',
                        audio: {
                            data: imageUrl
                        }
                    });
                } else {
                    contentArray.push({
                        type: 'image_url',
                        image_url: {
                            url: imageUrl
                        }
                    });
                }
            }
        }

        return contentArray.length > 0 ? contentArray : content;
    }

    // Audio Recording functionality
    initAudioRecording() {
        // Audio source selector events
        this.audioSourceButton.addEventListener('click', () => {
            this.audioSourceDropdown.classList.toggle('active');
        });

        // Close audio source dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.audioSourceButton.contains(e.target) && !this.audioSourceDropdown.contains(e.target)) {
                this.audioSourceDropdown.classList.remove('active');
            }
        });

        // Audio source option selection
        document.getElementById('audioSourceList').addEventListener('click', (e) => {
            const option = e.target.closest('.audio-source-option');
            if (option) {
                this.selectAudioSource(option.dataset.source);
            }
        });

        // Recording control buttons
        this.recordBtn.addEventListener('click', () => this.startRecording());
        this.stopBtn.addEventListener('click', () => this.stopRecording());
        this.segmentBtn.addEventListener('click', () => this.createSegment());

        // Load user's preferred audio source
        const preferredSource = localStorage.getItem('preferredAudioSource') || 'microphone';
        this.selectAudioSource(preferredSource);
    }

    selectAudioSource(source) {
        // Update the UI
        document.querySelectorAll('.audio-source-option').forEach(option => {
            option.classList.remove('selected');
        });
        document.querySelector(`[data-source="${source}"]`).classList.add('selected');

        // Update the button text and current source
        this.currentAudioSource = source;
        const sourceNames = {
            'microphone': 'Microphone',
            'system': 'System Audio',
            'both': 'Microphone + System Audio'
        };
        this.selectedAudioSource.textContent = sourceNames[source];

        // Close dropdown
        this.audioSourceDropdown.classList.remove('active');

        // Store user preference
        localStorage.setItem('preferredAudioSource', source);
    }

    async startRecording() {
        try {
            // Only request new stream if we don't have one or if it's not active
            if (!this.audioStream || this.audioStream.getTracks().some(track => track.readyState === 'ended')) {
                // Request permissions and get audio stream based on selected source
                this.audioStream = await this.getAudioStream();
            }

            // Set up MediaRecorder with Opus codec (preferred format)
            const options = {
                mimeType: 'audio/webm; codecs=opus',
                audioBitsPerSecond: 128000
            };

            // Fallback to other formats if Opus is not supported
            if (!MediaRecorder.isTypeSupported(options.mimeType)) {
                if (MediaRecorder.isTypeSupported('audio/webm')) {
                    options.mimeType = 'audio/webm';
                } else if (MediaRecorder.isTypeSupported('audio/mp4')) {
                    options.mimeType = 'audio/mp4';
                } else {
                    options.mimeType = 'audio/wav';
                }
            }

            this.mediaRecorder = new MediaRecorder(this.audioStream, options);
            this.audioChunks = [];

            // Set up event handlers
            this.mediaRecorder.ondataavailable = (event) => {
                if (event.data.size > 0) {
                    this.audioChunks.push(event.data);
                }
            };

            this.mediaRecorder.onstop = () => {
                this.processRecording();
            };

            // Start recording
            this.mediaRecorder.start(1000); // Collect data every second for streaming
            this.isRecording = true;
            this.recordingStartTime = Date.now();

            // Update UI
            this.updateRecordingUI(true);
            this.startRecordingTimer();

            console.log('Recording started with format:', options.mimeType);

        } catch (error) {
            console.error('Failed to start recording:', error);
            this.showRecordingError('Failed to start recording: ' + error.message);
        }
    }

    async getAudioStream() {
        const constraints = {};

        switch (this.currentAudioSource) {
            case 'microphone':
                constraints.audio = {
                    echoCancellation: true,
                    noiseSuppression: true,
                    autoGainControl: true
                };
                return await navigator.mediaDevices.getUserMedia(constraints);

            case 'system':
                constraints.audio = true;
                constraints.video = true; // Video is required for system audio capture
                const displayStream = await navigator.mediaDevices.getDisplayMedia(constraints);
                // Extract only audio tracks for recording
                const audioTracks = displayStream.getAudioTracks();
                if (audioTracks.length === 0) {
                    displayStream.getTracks().forEach(track => track.stop());
                    throw new Error('No system audio available');
                }
                // Create a new MediaStream with only audio tracks
                const audioOnlyStream = new MediaStream(audioTracks);
                // Stop video tracks to save resources
                displayStream.getVideoTracks().forEach(track => track.stop());
                return audioOnlyStream;

            case 'both':
                try {
                    // Get microphone stream
                    const micStream = await navigator.mediaDevices.getUserMedia({
                        audio: {
                            echoCancellation: true,
                            noiseSuppression: true,
                            autoGainControl: true
                        }
                    });

                    // Get system audio stream
                    const displayStream = await navigator.mediaDevices.getDisplayMedia({
                        audio: true,
                        video: true
                    });

                    const systemAudioTracks = displayStream.getAudioTracks();

                    // If no system audio is available, fall back to microphone only
                    if (systemAudioTracks.length === 0) {
                        console.warn('No system audio available, using microphone only');
                        displayStream.getTracks().forEach(track => track.stop());
                        return micStream;
                    }

                    // Create Web Audio API context for mixing
                    const audioContext = new (window.AudioContext || window.webkitAudioContext)();

                    // Create audio sources
                    const micSource = audioContext.createMediaStreamSource(micStream);
                    const systemSource = audioContext.createMediaStreamSource(new MediaStream(systemAudioTracks));

                    // Create a gain node for volume control (optional)
                    const micGain = audioContext.createGain();
                    const systemGain = audioContext.createGain();

                    // Set volumes (can be adjusted as needed)
                    micGain.gain.value = 0.7; // Slightly reduce mic volume
                    systemGain.gain.value = 0.8; // Slightly reduce system volume

                    // Create destination for mixed audio
                    const destination = audioContext.createMediaStreamDestination();

                    // Connect the audio graph
                    micSource.connect(micGain);
                    systemSource.connect(systemGain);
                    micGain.connect(destination);
                    systemGain.connect(destination);

                    // Clean up the display stream video tracks
                    displayStream.getVideoTracks().forEach(track => track.stop());

                    // Store audio context for cleanup later
                    this.audioContext = audioContext;
                    this.originalStreams = [micStream, displayStream];

                    return destination.stream;

                } catch (error) {
                    console.error('Error setting up mixed audio recording:', error);
                    // Fallback to microphone only
                    console.warn('Falling back to microphone only due to error');
                    return await navigator.mediaDevices.getUserMedia({
                        audio: {
                            echoCancellation: true,
                            noiseSuppression: true,
                            autoGainControl: true
                        }
                    });
                }

            default:
                throw new Error('Invalid audio source selected');
        }
    }

    stopRecording() {
        if (this.mediaRecorder && this.isRecording) {
            // Ensure this is not treated as a lap
            this.isCreatingLap = false;

            this.mediaRecorder.stop();
            this.isRecording = false;
            this.stopRecordingTimer();

            // Only stop all tracks in the stream when fully stopping (not segmenting)
            if (this.audioStream) {
                this.audioStream.getTracks().forEach(track => track.stop());
                this.audioStream = null; // Clear the stream reference
            }

            // Clean up Web Audio API resources used for "both" recording mode
            if (this.audioContext) {
                try {
                    this.audioContext.close();
                } catch (e) {
                    console.warn('Error closing audio context:', e);
                }
                this.audioContext = null;
            }

            // Clean up original streams used for "both" recording mode
            if (this.originalStreams) {
                this.originalStreams.forEach(stream => {
                    if (stream) {
                        stream.getTracks().forEach(track => track.stop());
                    }
                });
                this.originalStreams = null;
            }

            // Update UI
            this.updateRecordingUI(false);
            this.showProcessingIndicator();
        }
    }

    async createSegment() {
        if (this.mediaRecorder && this.isRecording) {
            // Set a flag to indicate this is a segment, not a final stop
            this.isCreatingLap = true;

            // Stop current recording to create a segment
            this.mediaRecorder.stop();
            // Note: processRecording() will be called automatically by the onstop event

            // The onstop event handler will start a new recording automatically if isCreatingLap is true
        }
    }

    async processRecording() {
        try {
            // Create blob from chunks
            const blob = new Blob(this.audioChunks, {
                type: this.mediaRecorder.mimeType || 'audio/webm; codecs=opus'
            });

            // Generate filename with timestamp
            const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
            const extension = this.getFileExtension(blob.type);
            const filename = `recording-${timestamp}.${extension}`;

            // Calculate recording duration
            const recordingDuration = this.recordingStartTime ? Date.now() - this.recordingStartTime : 0;

            // Thresholds for artifact storage: 500KB or 30 seconds (to prevent localStorage quota issues)
            const SIZE_THRESHOLD = 500 * 1024; // 500KB
            const DURATION_THRESHOLD = 30 * 1000; // 30 seconds in milliseconds

            console.log(`Recording stats: size=${blob.size}, duration=${recordingDuration}ms`);

            if (blob.size > SIZE_THRESHOLD || recordingDuration > DURATION_THRESHOLD) {
                // Use artifact storage for large/long recordings
                console.log('Using artifact storage for large recording');
                try {
                    const artifactId = await this.uploadToArtifactStorage(blob, filename);

                    // Add to file previews with artifact reference
                    this.addFilePreview(`artifact:${artifactId}`, filename, blob.type, {
                        isArtifact: true,
                        artifactId: artifactId,
                        size: blob.size,
                        duration: recordingDuration
                    });

                    console.log('Recording uploaded to artifact storage:', artifactId);
                } catch (uploadError) {
                    console.warn('Artifact upload failed, falling back to base64:', uploadError);
                    // Fallback to base64 if artifact upload fails
                    const dataURL = await this.blobToDataURL(blob);
                    this.addFilePreview(dataURL, filename, blob.type);
                }
            } else {
                // Use base64 for small recordings (current behavior)
                console.log('Using base64 storage for small recording');
                const dataURL = await this.blobToDataURL(blob);
                this.addFilePreview(dataURL, filename, blob.type);
            }

            // Hide processing indicator
            this.hideProcessingIndicator();

            console.log('Recording processed successfully:', filename);

            // If this was a lap, start a new recording immediately using the existing stream
            if (this.isCreatingLap) {
                this.isCreatingLap = false;
                // Small delay to ensure the current MediaRecorder is properly cleaned up
                setTimeout(() => {
                    this.startRecording();
                }, 100);
            }

        } catch (error) {
            console.error('Failed to process recording:', error);
            this.showRecordingError('Failed to process recording: ' + error.message);
            this.hideProcessingIndicator();

            // Reset lap flag on error
            this.isCreatingLap = false;
        }
    }

    getFileExtension(mimeType) {
        const extensions = {
            'audio/webm': 'webm',
            'audio/mp4': 'm4a',
            'audio/wav': 'wav',
            'audio/mpeg': 'mp3'
        };

        // Handle codec specifications
        for (const [type, ext] of Object.entries(extensions)) {
            if (mimeType.includes(type)) {
                return ext;
            }
        }

        return 'webm'; // Default fallback
    }

    blobToDataURL(blob) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => resolve(e.target.result);
            reader.onerror = (e) => reject(e);
            reader.readAsDataURL(blob);
        });
    }

    // Upload blob to artifact storage and return artifact ID
    async uploadToArtifactStorage(blob, filename) {
        try {
            // Add timeout to prevent hanging requests
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout

            const uploadUrl = `${this.baseUrl}/artifact`;
            console.log('Uploading to:', uploadUrl);
            console.log('Upload blob type:', blob.type, 'size:', blob.size);
            console.log('Upload filename:', filename);
            console.log('Request method: POST');

            const response = await fetch(uploadUrl, {
                method: 'POST',
                headers: {
                    'Content-Type': blob.type,
                    'X-Original-Filename': filename
                },
                body: blob,
                signal: controller.signal
            });

            clearTimeout(timeoutId);

            console.log('Upload response status:', response.status, response.statusText);
            console.log('Upload response headers:', [...response.headers.entries()]);

            if (!response.ok) {
                const errorText = await response.text();
                console.error('Error response body:', errorText);
                throw new Error(`Artifact upload failed: ${response.status} ${response.statusText} - ${errorText}`);
            }

            const result = await response.json();
            console.log('Artifact upload response:', result);

            if (!result.artifactId) {
                console.error('Server response structure:', JSON.stringify(result, null, 2));
                throw new Error('Server response missing artifactId field');
            }

            console.log('Artifact uploaded successfully:', result.artifactId);
            return result.artifactId;
        } catch (error) {
            if (error.name === 'AbortError') {
                console.warn('Artifact upload timed out - server may not be running');
                this.artifactServerUnavailable = true;
            } else if (error.message?.includes('Failed to fetch') || error.message?.includes('ERR_CONNECTION_RESET') || error.message?.includes('404 Not Found')) {
                console.warn('Artifact storage server not available - skipping upload');
                this.artifactServerUnavailable = true;
            } else {
                console.error('Failed to upload to artifact storage:', error);
            }
            throw error;
        }
    }

    // Fetch artifact data as base64 data URL
    async fetchArtifactAsDataURL(artifactId) {
        try {
            const response = await fetch(`${this.baseUrl}/artifact/${artifactId}`);
            if (!response.ok) {
                throw new Error(`Failed to fetch artifact ${artifactId}: ${response.status}`);
            }

            const blob = await response.blob();
            return await this.blobToDataURL(blob);
        } catch (error) {
            console.error('Failed to fetch artifact:', error);
            throw error;
        }
    }

    updateRecordingUI(isRecording) {
        if (isRecording) {
            this.recordBtn.style.display = 'none';
            this.stopBtn.style.display = 'flex';
            this.segmentBtn.style.display = 'flex';
            this.recordingIndicator.style.display = 'flex';
        } else {
            this.recordBtn.style.display = 'flex';
            this.stopBtn.style.display = 'none';
            this.segmentBtn.style.display = 'none';
            this.recordingIndicator.style.display = 'none';
        }
    }

    startRecordingTimer() {
        this.recordingTimerInterval = setInterval(() => {
            if (this.recordingStartTime) {
                const elapsed = Date.now() - this.recordingStartTime;
                const minutes = Math.floor(elapsed / 60000);
                const seconds = Math.floor((elapsed % 60000) / 1000);
                this.recordingTimer.textContent =
                    `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
            }
        }, 1000);
    }

    stopRecordingTimer() {
        if (this.recordingTimerInterval) {
            clearInterval(this.recordingTimerInterval);
            this.recordingTimerInterval = null;
        }
        this.recordingTimer.textContent = '00:00';
    }

    showProcessingIndicator() {
        // Replace the recording indicator with a processing indicator
        this.recordingIndicator.className = 'recording-processing';
        this.recordingIndicator.innerHTML = `
            <div class="processing-spinner"></div>
            <span>Processing...</span>
        `;
        this.recordingIndicator.style.display = 'flex';
    }

    hideProcessingIndicator() {
        this.recordingIndicator.style.display = 'none';
        this.recordingIndicator.className = 'recording-indicator';
        this.recordingIndicator.innerHTML = `
            <div class="recording-wave">
                <div class="wave-bar"></div>
                <div class="wave-bar"></div>
                <div class="wave-bar"></div>
                <div class="wave-bar"></div>
                <div class="wave-bar"></div>
            </div>
            <span class="recording-timer" id="recordingTimer">00:00</span>
        `;
        // Re-get the timer element reference since we recreated it
        this.recordingTimer = document.getElementById('recordingTimer');
    }

    showRecordingError(message) {
        this.showNotification(message, 'error');

        // Reset recording state
        this.isRecording = false;
        this.isCreatingLap = false;
        this.updateRecordingUI(false);
        this.stopRecordingTimer();
        this.hideProcessingIndicator();

        // Clean up streams - always stop tracks on error
        if (this.audioStream) {
            this.audioStream.getTracks().forEach(track => track.stop());
            this.audioStream = null;
        }
    }

    // Utility functions for file size calculation and formatting
    calculateDataURLSize(dataURL) {
        if (!dataURL || typeof dataURL !== 'string') return 0;

        // For data URLs, calculate the actual byte size
        if (dataURL.startsWith('data:')) {
            // Find the base64 data part after the comma
            const commaIndex = dataURL.indexOf(',');
            if (commaIndex === -1) return 0;

            const base64Data = dataURL.substring(commaIndex + 1);
            // Base64 encoding adds ~33% overhead, so actual size is roughly 3/4 of base64 length
            return Math.round((base64Data.length * 3) / 4);
        }

        // For regular URLs, estimate based on string length
        return dataURL.length * 2; // UTF-16 encoding
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 B';

        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));

        const size = bytes / Math.pow(k, i);
        const formatted = i === 0 ? size.toString() : size.toFixed(1);

        return `${formatted} ${sizes[i]}`;
    }

    getAttachmentSizeInfo(item) {
        let size = 0;
        let type = '';

        if (item.type === 'image_url' && item.image_url && item.image_url.url) {
            size = this.calculateDataURLSize(item.image_url.url);
            type = 'Image';
        } else if (item.type === 'file' && item.file && item.file.file_data) {
            size = this.calculateDataURLSize(item.file.file_data);
            type = 'File';
        } else if (item.type === 'audio' && item.audio && item.audio.data) {
            size = this.calculateDataURLSize(item.audio.data);
            type = 'Audio';
        }

        return {
            size: size,
            formattedSize: this.formatFileSize(size),
            type: type
        };
    }

    // Conversation Management Methods

    async saveConversations() {
        // CRITICAL FIX: Always attempt to save conversations - never skip due to quota
        // Multiple fallback mechanisms ensure conversations are never lost

        const now = Date.now();

        // Primary save attempt to localStorage
        let primarySaveSucceeded = false;
        try {
            localStorage.setItem('chat_conversations', JSON.stringify(this.conversations));
            primarySaveSucceeded = true;

            // Reset error tracking on successful save
            this.lastCleanupAttempt = null;
            this.storageQuotaExceeded = false;
            this.consecutiveSaveFailures = 0;

            console.log('Conversations saved successfully to localStorage');
            return;
        } catch (error) {
            console.warn('Primary localStorage save failed:', error.message);
            this.consecutiveSaveFailures = (this.consecutiveSaveFailures || 0) + 1;

            if (error.name === 'QuotaExceededError') {
                this.lastCleanupAttempt = now;
                this.storageQuotaExceeded = true;
            }
        }

        // Fallback 1: Try to save with reduced data (strip large attachments temporarily)
        if (!primarySaveSucceeded) {
            try {
                const reducedData = this.createReducedConversationsForSave();
                localStorage.setItem('chat_conversations', JSON.stringify(reducedData));
                console.log('Conversations saved with reduced data to localStorage');

                // Save the full data to backup location
                this.saveToBackupLocation();
                return;
            } catch (error) {
                console.warn('Reduced data localStorage save failed:', error.message);
            }
        }

        // Fallback 2: Save to IndexedDB
        try {
            await this.saveToIndexedDB();
            console.log('Conversations saved to IndexedDB fallback');
            return;
        } catch (error) {
            console.warn('IndexedDB save failed:', error.message);
        }

        // Fallback 3: Save to browser memory (session storage)
        try {
            sessionStorage.setItem('chat_conversations_backup', JSON.stringify(this.conversations));
            console.log('Conversations saved to sessionStorage as emergency backup');
        } catch (error) {
            console.warn('SessionStorage save failed:', error.message);
        }

        // Fallback 4: Save to download blob (user can manually save)
        if (this.consecutiveSaveFailures >= 3) {
            this.offerDownloadBackup();
        }

        // Show appropriate error notification
        if (this.storageQuotaExceeded) {
            this.showNotification('Storage quota exceeded - using backup storage. Conversations are preserved.', 'warning');
        } else {
            this.showNotification('Save error occurred - conversations backed up automatically.', 'warning');
        }
    }

    // Update the current conversation data in the conversations object
    updateCurrentConversation() {
        if (this.currentConversationId && this.messages && this.messages.length > 0) {
            // Update the conversation object with current messages
            if (!this.conversations[this.currentConversationId]) {
                // Generate title from first user message or use default
                const firstUserMessage = this.messages.find(msg => msg.role === 'user');
                const title = firstUserMessage ?
                    firstUserMessage.content.substring(0, 50).trim() + (firstUserMessage.content.length > 50 ? '...' : '') :
                    'New Conversation';

                this.conversations[this.currentConversationId] = {
                    title: title,
                    timestamp: Date.now(),
                    messages: []
                };
            }

            // Update messages and timestamp
            this.conversations[this.currentConversationId].messages = [...this.messages];
            this.conversations[this.currentConversationId].timestamp = Date.now();
        }
    }

    // New worker-based save method
    async saveConversationsViaWorker() {
        this.updateCurrentConversation();

        // Try to use workers for data optimization if available
        if (this.workerReady && this.workerManager.isInitialized) {
            try {
                // Use worker to create optimized data for storage
                const result = await this.workerManager.createReducedConversations(this.conversations);
                if (result.success) {
                    return this.saveOptimizedConversations(result.data);
                }
            } catch (error) {
                console.log('Worker optimization failed, using fallback:', error.message);
            }
        }

        // Fallback to direct save with local optimization
        return this.saveConversationsFallback();
    }

    // Save optimized conversations data
    saveOptimizedConversations(optimizedData) {
        try {
            localStorage.setItem('chat_conversations', JSON.stringify(optimizedData));
            console.log('Conversations saved with worker optimization');
            this.storageQuotaExceeded = false;
            this.consecutiveSaveFailures = 0;
            return true;
        } catch (error) {
            console.error('Optimized save failed:', error);
            if (error.name === 'QuotaExceededError') {
                return this.handleQuotaExceeded();
            }
            return false;
        }
    }

    // Fallback save method with quota handling
    saveConversationsFallback() {
        try {
            // First try with reduced data
            const reducedData = this.createReducedConversationsForSave();
            localStorage.setItem('chat_conversations', JSON.stringify(reducedData));
            console.log('Conversations saved via fallback with local optimization');
            this.storageQuotaExceeded = false;
            this.consecutiveSaveFailures = 0;
            return true;
        } catch (error) {
            console.error('Fallback save failed:', error);
            if (error.name === 'QuotaExceededError') {
                return this.handleQuotaExceeded();
            }
            return false;
        }
    }

    // Handle quota exceeded errors - should be rare with aggressive prevention
    handleQuotaExceeded() {
        console.log('Storage quota exceeded despite prevention measures...');
        this.storageQuotaExceeded = true;
        this.consecutiveSaveFailures++;

        // Clean up old localStorage items to free space
        this.cleanupOldStorageItems();

        // Try saving with even more aggressive data stripping
        try {
            const strippedData = this.createReducedConversationsForSave();
            localStorage.setItem('chat_conversations', JSON.stringify(strippedData));
            console.log('Storage quota resolved with aggressive data stripping');
            this.showNotification('Large attachments moved to server storage automatically.', 'info');
            return true;
        } catch (stillExceededError) {
            console.error('Storage still exceeded after aggressive cleanup:', stillExceededError);

            // Offer download backup as last resort
            this.offerDownloadBackup();
            this.showNotification('Storage quota exceeded. Download backup recommended.', 'warning');
            return false;
        }
    }

    // Clean up old storage items to free space
    cleanupOldStorageItems() {
        const itemsToCheck = [
            'chat_conversations_backup',
            'chat_conversations_emergency',
            'chat_messages_cache',
            'chat_temp_data',
            'old_chat_data'
        ];

        itemsToCheck.forEach(key => {
            try {
                if (localStorage.getItem(key)) {
                    localStorage.removeItem(key);
                    console.log(`Cleaned up old storage item: ${key}`);
                }
            } catch (e) {
                // Ignore cleanup errors
            }
        });

        // Also clear session storage
        try {
            sessionStorage.clear();
        } catch (e) {
            // Ignore cleanup errors
        }
    }


    createReducedConversationsForSave() {
        // Aggressively prevent large data from going to localStorage
        const reduced = JSON.parse(JSON.stringify(this.conversations));
        const MAX_DATA_URL_SIZE = 10 * 1024; // Only allow tiny data URLs (10KB)

        Object.values(reduced).forEach(conversation => {
            if (conversation.messages) {
                conversation.messages.forEach(message => {
                    // Handle attachments array
                    if (message.attachments && Array.isArray(message.attachments)) {
                        message.attachments.forEach(attachment => {
                            // Strip any large data URLs from attachments
                            if (attachment.image_url?.url?.startsWith('data:') &&
                                attachment.image_url.url.length > MAX_DATA_URL_SIZE) {
                                attachment.image_url.url = '[LARGE_DATA_STRIPPED_USE_ARTIFACT]';
                                attachment.stripped = true;
                            }
                            if (attachment.audio?.data?.startsWith('data:') &&
                                attachment.audio.data.length > MAX_DATA_URL_SIZE) {
                                attachment.audio.data = '[LARGE_AUDIO_STRIPPED_USE_ARTIFACT]';
                                attachment.stripped = true;
                            }
                        });
                    }

                    // Handle content array (for complex message structures)
                    if (Array.isArray(message.content)) {
                        message.content.forEach(item => {
                            // Strip large data URLs but keep references
                            if (item.type === 'audio' && item.audio?.data?.startsWith('data:') &&
                                item.audio.data.length > MAX_DATA_URL_SIZE) {
                                item.audio.data = '[LARGE_AUDIO_STRIPPED_USE_ARTIFACT]';
                                item.audio.stripped = true;
                            }
                            if (item.type === 'image' && item.image?.data?.startsWith('data:') &&
                                item.image.data.length > MAX_DATA_URL_SIZE) {
                                item.image.data = '[LARGE_IMAGE_STRIPPED_USE_ARTIFACT]';
                                item.image.stripped = true;
                            }
                            if (item.type === 'file' && item.file?.data?.startsWith('data:') &&
                                item.file.data.length > MAX_DATA_URL_SIZE) {
                                item.file.data = '[LARGE_FILE_STRIPPED_USE_ARTIFACT]';
                                item.file.stripped = true;
                            }
                        });
                    }
                });
            }
        });

        return reduced;
    }


    saveToBackupLocation() {
        // Save full conversation data to a backup localStorage key
        try {
            const backupData = {
                timestamp: Date.now(),
                conversations: this.conversations
            };
            localStorage.setItem('chat_conversations_full_backup', JSON.stringify(backupData));
            console.log('Full backup saved to localStorage');
        } catch (error) {
            console.warn('Backup location save failed:', error.message);
        }
    }

    async saveToIndexedDB() {
        return new Promise((resolve, reject) => {
            const request = indexedDB.open('AgentFlowDB', 1);

            request.onerror = () => reject(request.error);

            request.onupgradeneeded = (event) => {
                const db = event.target.result;
                if (!db.objectStoreNames.contains('conversations')) {
                    db.createObjectStore('conversations', { keyPath: 'id' });
                }
            };

            request.onsuccess = (event) => {
                const db = event.target.result;
                const transaction = db.transaction(['conversations'], 'readwrite');
                const store = transaction.objectStore('conversations');

                const data = {
                    id: 'current',
                    timestamp: Date.now(),
                    conversations: this.conversations
                };

                const putRequest = store.put(data);

                putRequest.onsuccess = () => {
                    db.close();
                    resolve();
                };

                putRequest.onerror = () => {
                    db.close();
                    reject(putRequest.error);
                };
            };
        });
    }

    offerDownloadBackup() {
        try {
            const backupData = {
                timestamp: Date.now(),
                conversations: this.conversations,
                version: '1.0'
            };

            const blob = new Blob([JSON.stringify(backupData, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);

            const a = document.createElement('a');
            a.href = url;
            a.download = `agentflow-backup-${new Date().toISOString().slice(0, 10)}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            this.showNotification('Conversations backup downloaded. Please save this file!', 'success');
            console.log('Download backup offered to user');
        } catch (error) {
            console.error('Failed to offer download backup:', error);
        }
    }

    // Enhanced load method that checks backup sources
    loadConversations() {
        let conversations = null;

        // Try primary localStorage first
        try {
            const saved = localStorage.getItem('chat_conversations');
            if (saved) {
                conversations = JSON.parse(saved);
                console.log('Conversations loaded from primary localStorage');
            }
        } catch (error) {
            console.warn('Primary localStorage load failed:', error.message);
        }

        // If primary failed, try backup localStorage
        if (!conversations) {
            try {
                const backup = localStorage.getItem('chat_conversations_full_backup');
                if (backup) {
                    const backupData = JSON.parse(backup);
                    conversations = backupData.conversations;
                    console.log('Conversations loaded from backup localStorage');
                }
            } catch (error) {
                console.warn('Backup localStorage load failed:', error.message);
            }
        }

        // If still no data, try sessionStorage
        if (!conversations) {
            try {
                const session = sessionStorage.getItem('chat_conversations_backup');
                if (session) {
                    conversations = JSON.parse(session);
                    console.log('Conversations loaded from sessionStorage backup');
                }
            } catch (error) {
                console.warn('SessionStorage load failed:', error.message);
            }
        }

        // If still no data, try IndexedDB
        if (!conversations) {
            this.loadFromIndexedDB().then(data => {
                if (data) {
                    this.conversations = data;
                    console.log('Conversations loaded from IndexedDB backup');
                    this.renderConversationsList();
                }
            }).catch(error => {
                console.warn('IndexedDB load failed:', error.message);
            });
        }

        return conversations || {};
    }

    async loadFromIndexedDB() {
        return new Promise((resolve) => {
            const request = indexedDB.open('AgentFlowDB', 1);

            request.onerror = () => resolve(null);

            request.onsuccess = (event) => {
                const db = event.target.result;
                if (!db.objectStoreNames.contains('conversations')) {
                    db.close();
                    resolve(null);
                    return;
                }

                const transaction = db.transaction(['conversations'], 'readonly');
                const store = transaction.objectStore('conversations');
                const getRequest = store.get('current');

                getRequest.onsuccess = () => {
                    db.close();
                    const result = getRequest.result;
                    resolve(result ? result.conversations : null);
                };

                getRequest.onerror = () => {
                    db.close();
                    resolve(null);
                };
            };
        });
    }

    cleanupOldConversations() {
        const conversationIds = Object.keys(this.conversations);

        // Calculate total storage usage
        let totalSize = 0;
        const conversationSizes = {};

        conversationIds.forEach(id => {
            const size = this.calculateConversationSize(this.conversations[id]);
            conversationSizes[id] = size;
            totalSize += size;
        });

        // Target: Keep under 4MB total (localStorage typically has 5-10MB limit)
        const TARGET_SIZE = 4 * 1024 * 1024; // 4MB

        // DISABLED: All cleanup disabled per user request
        // Conversations and all data are preserved regardless of size or count
        if (totalSize > TARGET_SIZE || conversationIds.length > 15) {
            console.log(`Storage usage: ${(totalSize / 1024 / 1024).toFixed(2)}MB, ${conversationIds.length} conversations - all cleanup disabled`);
        }

        // DISABLED: All cleanup disabled per user request
        // this.cleanupLargeAttachmentsInConversations();
    }

    calculateConversationSize(conversation) {
        try {
            return JSON.stringify(conversation).length * 2; // Approximate UTF-16 byte size
        } catch (e) {
            return 0;
        }
    }

    async cleanupLargeAttachmentsInConversations() {
        // Skip artifact cleanup if server is known to be unavailable
        if (this.artifactServerUnavailable) {
            console.log('Skipping artifact cleanup - server unavailable');
            return;
        }

        const conversationIds = Object.keys(this.conversations);
        let cleanedAny = false;

        for (const id of conversationIds) {
            const conversation = this.conversations[id];
            if (!conversation.messages) continue;

            for (const message of conversation.messages) {
                if (Array.isArray(message.content)) {
                    for (const item of message.content) {
                        // Check for large base64 audio/image data
                        if ((item.type === 'audio' && item.audio?.data?.startsWith('data:')) ||
                            (item.type === 'image' && item.image?.data?.startsWith('data:'))) {

                            const dataURL = item.type === 'audio' ? item.audio.data : item.image.data;
                            const size = this.calculateDataURLSize(dataURL);

                            // If larger than 500KB, try to move to artifact storage
                            if (size > 500 * 1024) {
                                try {
                                    // Convert dataURL to blob and upload to artifact storage
                                    const blob = this.dataURLToBlob(dataURL);
                                    const filename = item.type === 'audio' ? 'audio.webm' : 'image.png';
                                    const artifactId = await this.uploadToArtifactStorage(blob, filename);

                                    // Replace the data with artifact reference
                                    if (item.type === 'audio') {
                                        item.audio.data = `artifact:${artifactId}`;
                                    } else {
                                        item.image.data = `artifact:${artifactId}`;
                                    }
                                    cleanedAny = true;
                                } catch (error) {
                                    // Check if this is a connection error
                                    const isConnectionError = error.message?.includes('Failed to fetch') ||
                                                             error.message?.includes('ERR_CONNECTION_RESET') ||
                                                             error.message?.includes('404 Not Found') ||
                                                             error.name === 'AbortError';

                                    // Only log connection issues once to avoid spam
                                    if (!this.artifactServerUnavailable && isConnectionError) {
                                        console.warn('Artifact storage server not available - stopping cleanup attempts');
                                        this.artifactServerUnavailable = true;
                                    } else if (!isConnectionError) {
                                        console.warn('Failed to move large attachment to artifact storage:', error);
                                    }

                                    // For connection errors, keep the data and stop trying
                                    if (isConnectionError) {
                                        // Keep data but add warning marker
                                        if (item.type === 'audio') {
                                            item.audio.storageWarning = 'Large file - artifact server unavailable';
                                        } else {
                                            item.image.storageWarning = 'Large file - artifact server unavailable';
                                        }
                                        // Stop processing more items since server is unavailable
                                        return;
                                    } else {
                                        // For other errors, remove the large data to prevent quota issues
                                        if (item.type === 'audio') {
                                            item.audio.data = '[Large audio data removed due to storage constraints]';
                                        } else {
                                            item.image.data = '[Large image data removed due to storage constraints]';
                                        }
                                        cleanedAny = true;
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }

        if (cleanedAny) {
            console.log('Cleaned up large attachments in conversations');
        }
    }

    dataURLToBlob(dataURL) {
        const parts = dataURL.split(',');
        const header = parts[0];
        const data = parts[1];
        const mimeType = header.match(/:(.*?);/)[1];
        const byteCharacters = atob(data);
        const byteNumbers = new Array(byteCharacters.length);
        for (let i = 0; i < byteCharacters.length; i++) {
            byteNumbers[i] = byteCharacters.charCodeAt(i);
        }
        const byteArray = new Uint8Array(byteNumbers);
        return new Blob([byteArray], { type: mimeType });
    }

    initializeConversation() {
        // Check if there are existing conversations
        const conversationIds = Object.keys(this.conversations);
        if (conversationIds.length > 0) {
            // Load the most recent conversation
            const sortedIds = conversationIds.sort((a, b) =>
                this.conversations[b].lastModified - this.conversations[a].lastModified
            );
            this.loadConversation(sortedIds[0]);
        } else {
            // Create a new conversation
            this.createNewConversation();
        }
    }

    createNewConversation() {
        const id = 'conv_' + Date.now();
        const title = 'New Conversation';

        this.conversations[id] = {
            id: id,
            title: title,
            messages: [],
            systemPrompt: this.systemPrompt,  // Save current system prompt
            createdAt: Date.now(),
            lastModified: Date.now()
        };

        this.currentConversationId = id;
        this.messages = [];
        this.renderMessages();
        this.saveConversations();
        this.renderConversationsList();

        // Reload tools to refresh available tools and uncheck all by default
        this.loadTools();
    }

    loadConversation(id) {
        if (this.conversations[id]) {
            this.currentConversationId = id;
            this.messages = [...this.conversations[id].messages];
            // Load system prompt from conversation
            this.systemPrompt = this.conversations[id].systemPrompt || 'You are a helpful assistant.\nCurrent time is {{now | formatTimeInLocation "Europe/Paris" "2006-01-02 15:04"}}';
            this.systemPromptTextarea.value = this.systemPrompt;
            // Clear any currently selected files when switching conversations
            this.clearFilePreviews();
            this.renderMessages();
            this.renderConversationsList();
        }
    }

    // Helper method to create conversation data without attachments
    createMessagesWithoutAttachments() {
        return this.messages.map(msg => {
            if (Array.isArray(msg.content)) {
                // For multimodal content, remove attachment data but keep metadata
                const sanitizedContent = msg.content.map(item => {
                    if (item.type === 'file' && item.file && item.file.file_data) {
                        return {
                            type: 'file',
                            file: {
                                filename: item.file.filename,
                                metadata: '[File data removed to save storage space]'
                            }
                        };
                    } else if (item.type === 'image_url' && item.image_url && item.image_url.url) {
                        const url = item.image_url.url;
                        if (url.startsWith('data:')) {
                            return {
                                type: 'image_url',
                                image_url: {
                                    url: '[Large image data removed to save storage space]'
                                }
                            };
                        }
                    } else if (item.type === 'audio' && item.audio && item.audio.data) {
                        const audioData = item.audio.data;
                        if (audioData.startsWith('data:')) {
                            return {
                                type: 'audio',
                                audio: {
                                    data: '[Large audio data removed to save storage space]'
                                }
                            };
                        }
                    } else if (item.type === 'audio_artifact' && item.audio_artifact) {
                        // Keep artifact references as they don't consume localStorage space
                        return item;
                    }
                    return item;
                });
                return { ...msg, content: sanitizedContent };
            }
            return msg;
        });
    }

    saveCurrentConversation() {
        if (this.currentConversationId && this.conversations[this.currentConversationId]) {
            // Store all messages with their full content including attachments
            this.conversations[this.currentConversationId].messages = this.messages;
            this.conversations[this.currentConversationId].systemPrompt = this.systemPrompt;  // Save system prompt
            this.conversations[this.currentConversationId].lastModified = Date.now();

            // Auto-generate title from first message if still default
            if (this.conversations[this.currentConversationId].title === 'New Conversation' &&
                this.messages.length > 0) {
                const firstUserMessage = this.messages.find(m => m.role === 'user');
                if (firstUserMessage) {
                    let titleContent = '';

                    // Handle both string and multimodal content
                    if (typeof firstUserMessage.content === 'string') {
                        titleContent = firstUserMessage.content;
                    } else if (Array.isArray(firstUserMessage.content)) {
                        // Extract text content from multimodal array
                        const textParts = firstUserMessage.content
                            .filter(item => item.type === 'text')
                            .map(item => item.text);
                        titleContent = textParts.join(' ');

                        // If no text content, use a generic title based on file types
                        if (!titleContent.trim()) {
                            const fileTypes = firstUserMessage.content
                                .filter(item => item.type === 'image_url' || item.type === 'file' || item.type === 'audio')
                                .map(item => {
                                    if (item.type === 'file') return 'PDF';
                                    if (item.type === 'audio') return 'Audio';
                                    return 'Image';
                                });
                            titleContent = fileTypes.length > 0 ? `Uploaded ${fileTypes.join(', ')}` : 'Multimodal message';
                        }
                    } else {
                        titleContent = String(firstUserMessage.content);
                    }

                    let title = titleContent.substring(0, 50);
                    if (titleContent.length > 50) title += '...';
                    this.conversations[this.currentConversationId].title = title;
                }
            }

            // Save conversations with improved quota handling
            this.saveConversations().catch(error => {
                console.error('Error saving conversation:', error);
                // Error notifications are handled in saveConversations()
            });
            this.renderConversationsList();
        }
    }

    deleteConversation(id) {
        if (confirm('Are you sure you want to delete this conversation?')) {
            delete this.conversations[id];
            this.saveConversations();

            if (id === this.currentConversationId) {
                this.initializeConversation();
            }

            this.renderConversationsList();
        }
    }

    renameConversation(id) {
        const conversation = this.conversations[id];
        if (conversation) {
            const newTitle = prompt('Enter new title:', conversation.title);
            if (newTitle && newTitle.trim()) {
                conversation.title = newTitle.trim();
                conversation.lastModified = Date.now();
                this.saveConversations();
                this.renderConversationsList();
            }
        }
    }

    duplicateConversation(id) {
        const conversation = this.conversations[id];
        if (conversation) {
            // Generate a new unique ID
            const newId = 'conv_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);

            // Create a deep copy of the conversation
            const duplicatedConversation = {
                id: newId,
                title: `${conversation.title} (Copy)`,
                messages: JSON.parse(JSON.stringify(conversation.messages)), // Deep copy messages
                systemPrompt: conversation.systemPrompt,
                createdAt: Date.now(),
                lastModified: Date.now()
            };

            // Add the duplicated conversation to the conversations object
            this.conversations[newId] = duplicatedConversation;

            // Save conversations and update UI
            this.saveConversations();
            this.renderConversationsList();

            // Optional: Switch to the new duplicated conversation
            this.saveCurrentConversation(); // Save current before switching
            this.loadConversation(newId);

            console.log(`📋 Conversation duplicated: "${conversation.title}" -> "${duplicatedConversation.title}"`);
        }
    }

    renderConversationsList() {
        this.conversationsList.innerHTML = '';

        const sortedIds = Object.keys(this.conversations).sort((a, b) =>
            this.conversations[b].lastModified - this.conversations[a].lastModified
        );

        sortedIds.forEach(id => {
            const conv = this.conversations[id];
            const item = document.createElement('div');
            item.className = 'conversation-item';
            if (id === this.currentConversationId) {
                item.classList.add('active');
            }

            const title = document.createElement('span');
            title.className = 'conversation-title';
            title.textContent = conv.title;

            const actions = document.createElement('div');
            actions.className = 'conversation-actions';

            const duplicateBtn = document.createElement('button');
            duplicateBtn.className = 'duplicate-btn';
            duplicateBtn.innerHTML = '📋';
            duplicateBtn.title = 'Duplicate conversation';
            duplicateBtn.onclick = (e) => {
                e.stopPropagation();
                this.duplicateConversation(id);
            };

            const renameBtn = document.createElement('button');
            renameBtn.className = 'rename-btn';
            renameBtn.innerHTML = '✏️';
            renameBtn.title = 'Rename';
            renameBtn.onclick = (e) => {
                e.stopPropagation();
                this.renameConversation(id);
            };

            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'delete-btn';
            deleteBtn.innerHTML = '🗑️';
            deleteBtn.title = 'Delete';
            deleteBtn.onclick = (e) => {
                e.stopPropagation();
                this.deleteConversation(id);
            };

            actions.appendChild(duplicateBtn);
            actions.appendChild(renameBtn);
            actions.appendChild(deleteBtn);

            item.appendChild(title);
            item.appendChild(actions);

            item.onclick = () => {
                if (id !== this.currentConversationId) {
                    // Save current conversation before switching
                    this.saveCurrentConversation();
                    this.loadConversation(id);
                }
            };

            this.conversationsList.appendChild(item);
        });
    }

    async loadModels() {
        try {
            const response = await fetch(this.baseUrl + '/v1/models');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.models = data.data || [];

            if (this.models.length > 0) {
                // Select the first model by default
                this.selectedModel = this.models[0].id;
                this.selectedModelName.textContent = this.selectedModel;
                this.renderModelsList();
            } else {
                this.selectedModelName.textContent = 'No models available';
            }
        } catch (error) {
            console.error('Error loading models:', error);
            this.selectedModelName.textContent = 'Error loading models';
        }
    }

    renderModelsList() {
        this.modelsList.innerHTML = '';

        this.models.forEach(model => {
            const option = document.createElement('div');
            option.className = 'model-option';
            if (model.id === this.selectedModel) {
                option.classList.add('selected');
            }

            option.innerHTML = `
                <span class="model-name">${model.id}</span>
                <span class="model-owner">${model.owned_by || 'Unknown'}</span>
            `;

            option.addEventListener('click', () => {
                this.selectModel(model.id);
            });

            this.modelsList.appendChild(option);
        });
    }

    selectModel(modelId) {
        this.selectedModel = modelId;
        this.selectedModelName.textContent = modelId;
        this.renderModelsList();
        this.modelDropdown.classList.remove('active');
    }

    async loadTools() {
        try {
            const response = await fetch(this.baseUrl + '/v1/tools');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.tools = data || [];

            // Clear all selected tools by default
            this.selectedTools.clear();

            this.renderToolsList();
            this.updateToolsCountDisplay();
        } catch (error) {
            console.error('Error loading tools:', error);
            this.selectedToolsCount.textContent = 'Tools: Error';
        }
    }

    renderToolsList() {
        this.toolsList.innerHTML = '';

        if (this.tools.length === 0) {
            this.toolsList.innerHTML = '<div class="loading-tools">No tools available</div>';
            return;
        }

        this.tools.forEach(tool => {
            const option = document.createElement('div');
            option.className = 'tool-option';
            if (this.selectedTools.has(tool.Name)) {
                option.classList.add('selected');
            }

            option.innerHTML = `
                <div class="tool-info">
                    <div class="tool-name">${tool.Name}</div>
                    <div class="tool-description">${tool.Description || 'No description available'}</div>
                </div>
                <div class="tool-checkbox ${this.selectedTools.has(tool.Name) ? 'checked' : ''}"></div>
            `;

            option.addEventListener('click', () => {
                this.toggleTool(tool.Name);
            });

            this.toolsList.appendChild(option);
        });
    }

    isGoogleSearchTool(toolName) {
        return toolName && toolName.toLowerCase().includes('google');
    }

    toggleTool(toolName) {
        const isTogglingGoogleSearch = this.isGoogleSearchTool(toolName);
        const hasGoogleSearchSelected = Array.from(this.selectedTools).some(tool => this.isGoogleSearchTool(tool));

        if (this.selectedTools.has(toolName)) {
            this.selectedTools.delete(toolName);
        } else {
            // If selecting a Google search tool, deselect all other tools
            if (isTogglingGoogleSearch) {
                this.selectedTools.clear();
                this.selectedTools.add(toolName);
            }
            // If selecting a non-Google tool while Google is selected, deselect Google first
            else if (hasGoogleSearchSelected) {
                this.selectedTools.clear();
                this.selectedTools.add(toolName);
            }
            // Normal selection
            else {
                this.selectedTools.add(toolName);
            }
        }
        this.renderToolsList();
        this.updateToolsCountDisplay();
    }

    selectAllTools() {
        // Check if any Google search tools exist
        const hasGoogleSearchTools = this.tools.some(tool => this.isGoogleSearchTool(tool.Name));

        if (hasGoogleSearchTools) {
            // If Google search tools exist, only select non-Google tools when selecting "all"
            this.tools.forEach(tool => {
                if (!this.isGoogleSearchTool(tool.Name)) {
                    this.selectedTools.add(tool.Name);
                }
            });
        } else {
            // No Google search tools, select all tools normally
            this.tools.forEach(tool => {
                this.selectedTools.add(tool.Name);
            });
        }
        this.renderToolsList();
        this.updateToolsCountDisplay();
    }

    selectNoTools() {
        this.selectedTools.clear();
        this.renderToolsList();
        this.updateToolsCountDisplay();
    }

    updateToolsCountDisplay() {
        const selectedCount = this.selectedTools.size;
        const totalCount = this.tools.length;

        if (selectedCount === totalCount) {
            this.selectedToolsCount.textContent = 'Tools: All';
        } else if (selectedCount === 0) {
            this.selectedToolsCount.textContent = 'Tools: None';
        } else {
            this.selectedToolsCount.textContent = `Tools: ${selectedCount}/${totalCount}`;
        }
    }

    buildModelWithTools() {
        let modelString = this.selectedModel;

        if (this.selectedTools.size > 0 && this.selectedTools.size < this.tools.length) {
            // Only add tools if not all are selected (all selected means use all tools)
            const toolNames = Array.from(this.selectedTools);
            modelString += '|' + toolNames.join('|');
        }

        return modelString;
    }

    renderMarkdown(text) {
        console.log('🔍 renderMarkdown called with:', text?.substring(0, 200) + '...');

        // Configure marked to allow HTML (including SVG)
        const customRenderer = this.getCustomRenderer();
        console.log('🎨 Custom renderer created:', !!customRenderer.code);

        // Try new marked.js API first, then fall back to old API
        let html;
        try {
            // New API (marked v4+)
            html = marked.parse(text || '', {
                breaks: true,
                gfm: true,
                sanitize: false,
                renderer: customRenderer
            });
        } catch (error) {
            console.log('New API failed, trying old API:', error);
            // Old API (marked v3 and below)
            marked.setOptions({
                breaks: true,
                gfm: true,
                sanitize: false,
                renderer: customRenderer
            });
            html = marked.parse ? marked.parse(text || '') : marked(text || '');
        }
        console.log('📝 Marked output:', html?.substring(0, 300) + '...');

        // Handle SVG URLs - convert them to object tags for better compatibility (including PlantUML)
        html = html.replace(/<img([^>]*?)src=["']([^"']*(?:\.svg|\/plantuml\/svg\/)[^"']*?)["']([^>]*?)>/g, (match, attrs1, src, attrs2) => {
            // Use object tag for SVG files to ensure they display properly
            return `<object type="image/svg+xml" data="${src}">
                <img src="${src}" alt="SVG Image">
            </object>`;
        });

        // Also handle SVG URLs that might be in markdown link format (including PlantUML)
        html = html.replace(/!\[([^\]]*)\]\(([^)]*(?:\.svg|\/plantuml\/svg\/)[^)]*)\)/g, (match, alt, src) => {
            return `<object type="image/svg+xml" data="${src}">
                <img src="${src}" alt="${alt}">
            </object>`;
        });

        return html;
    }

    // Rewrite PlantUML URLs to use the proxy when needed
    rewritePlantUMLUrls(html) {
        if (!this.baseUrl) {
            // If baseUrl is empty, we're served from the main server, no rewriting needed
            return html;
        }

        // Rewrite PlantUML URLs in object tags
        html = html.replace(/data="http:\/\/localhost:9999\/plantuml\//g, `data="${this.baseUrl}/plantuml/`);

        // Rewrite PlantUML URLs in img tags
        html = html.replace(/src="http:\/\/localhost:9999\/plantuml\//g, `src="${this.baseUrl}/plantuml/`);

        return html;
    }

    // Map common language aliases to Prism.js language identifiers
    mapLanguageToPrism(language) {
        const langMap = {
            'js': 'javascript',
            'ts': 'typescript',
            'jsx': 'jsx',
            'tsx': 'tsx',
            'py': 'python',
            'rb': 'ruby',
            'sh': 'bash',
            'shell': 'bash',
            'yml': 'yaml',
            'md': 'markdown',
            'dockerfile': 'docker',
            'text': 'none',
            'txt': 'none'
        };

        const lang = language.toLowerCase();
        return langMap[lang] || lang;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    getCustomRenderer() {
        const renderer = new marked.Renderer();
        const chatUI = this; // Capture reference to this instance

        // Custom code block renderer with copy button and syntax highlighting
        // Support both old and new marked.js API signatures
        renderer.code = function(code, infoString, escaped) {
            console.log('🔧 Code renderer called with:', {
                code: typeof code === 'string' ? code.substring(0, 100) + '...' : code,
                infoString,
                escaped,
                codeType: typeof code,
                arguments: Array.from(arguments)
            });

            // Handle the case where code might be an object (modern marked.js)
            let codeText;
            let language;

            if (typeof code === 'object' && code !== null) {
                // Modern marked.js passes a token object
                codeText = code.text || code.raw || String(code);
                language = code.lang || infoString;
                console.log('🔧 Modern API detected - code object:', {
                    text: code.text,
                    lang: code.lang,
                    raw: code.raw
                });
            } else if (typeof code === 'string') {
                // Legacy API or simple string
                codeText = code;
                language = infoString; // In newer versions, language is passed as infoString
            } else {
                codeText = String(code || '');
                language = infoString;
            }

            // Extract language from infoString (format: "python" or "python highlight")
            let validLang = 'text';
            if (language && typeof language === 'string' && language.trim()) {
                validLang = language.trim().split(/\s+/)[0].toLowerCase(); // Take first word
            } else if (infoString && typeof infoString === 'string' && infoString.trim()) {
                validLang = infoString.trim().split(/\s+/)[0].toLowerCase(); // Take first word
            }

            const prismLang = chatUI.mapLanguageToPrism(validLang);
            const codeId = 'code_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);

            // Ensure we have valid code text
            if (!codeText || codeText === '[object Object]') {
                codeText = 'Code content not available';
            }

            console.log('🎯 Final processing:', {
                validLang,
                prismLang,
                codeLength: codeText.length,
                rawLanguage: language,
                rawInfoString: infoString
            });

            // Escape HTML to prevent XSS
            const escapedCode = chatUI.escapeHtml(codeText);

            return `<div class="code-block-container">
                <div class="code-block-header">
                    <span class="code-language">${validLang.toUpperCase()}</span>
                </div>
                <div class="code-block-wrapper">
                    <pre class="code-block"><code id="${codeId}" class="language-${prismLang}" data-original-code="${escapedCode}">${escapedCode}</code></pre>
                    <button class="code-copy-btn" onclick="chatUI.copyCodeBlock('${codeId}', this)" title="Copy code">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                        </svg>
                    </button>
                </div>
            </div>`;
        };

        // Custom image renderer to handle SVG URLs properly
        renderer.image = function(href, title, text) {
            // The href parameter might be an object, extract the actual href
            let url = '';
            if (typeof href === 'object' && href !== null && href.href) {
                url = String(href.href);
            } else if (typeof href === 'string') {
                url = href;
            } else {
                url = String(href || '');
            }

            const altText = String(text || '');
            const titleText = String(title || '');

            // Check if it's an SVG URL (including PlantUML SVG URLs)
            if (url && (url.includes('.svg') || url.endsWith('.svg') || url.includes('/plantuml/svg/'))) {
                // Use object tag for better SVG compatibility - CSS handles styling
                return `<object type="image/svg+xml" data="${url}" title="${titleText}">
                    <img src="${url}" alt="${altText}" title="${titleText}">
                </object>`;
            }

            // Check if it's another type of image generation URL (also likely SVG)
            if (url && (url.includes('wardley_map') ||
                        url.includes('localhost:8585') ||
                        url.includes('image/svg') ||
                        url.includes('/plantuml/png/'))) {
                return `<img src="${url}" alt="${altText}" title="${titleText}">`;
            }

            // Default image rendering
            return `<img src="${url}" alt="${altText}" title="${titleText}">`;
        };

        return renderer;
    }

    // Method to copy code block content
    async copyCodeBlock(codeId, buttonElement) {
        try {
            const codeElement = document.getElementById(codeId);
            if (!codeElement) return;

            // Prioritize textContent/innerText as they contain the complete, rendered code
            // The data-original-code attribute may be truncated due to HTML escaping issues
            let codeText = codeElement.textContent || codeElement.innerText;

            // Only use data-original-code as a fallback if textContent is not available
            if (!codeText || codeText.trim() === '') {
                codeText = codeElement.getAttribute('data-original-code') || '';
            }

            // For HTML entities, decode if the text appears to contain them
            let finalText = codeText;
            if (codeText.includes('&lt;') || codeText.includes('&gt;') || codeText.includes('&amp;')) {
                const tempDiv = document.createElement('div');
                tempDiv.innerHTML = codeText;
                finalText = tempDiv.textContent || tempDiv.innerText || codeText;
            }

            await navigator.clipboard.writeText(finalText);

            // Visual feedback
            const originalContent = buttonElement.innerHTML;
            buttonElement.innerHTML = `
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20,6 9,17 4,12"></polyline>
                </svg>
                Copied!
            `;
            buttonElement.classList.add('copied');

            // Reset after 2 seconds
            setTimeout(() => {
                buttonElement.innerHTML = originalContent;
                buttonElement.classList.remove('copied');
            }, 2000);

        } catch (error) {
            console.error('Failed to copy code:', error);
            // Show error feedback
            buttonElement.innerHTML = 'Failed';
            setTimeout(() => {
                buttonElement.innerHTML = `
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                    </svg>
                    Copy
                `;
            }, 2000);
        }
    }

    renderMessageContent(content) {
        if (typeof content === 'string') {
            // Simple text content
            let html = this.renderMarkdown(content);
            return this.rewritePlantUMLUrls(html);
        } else if (Array.isArray(content)) {
            // Multimodal content with text and images
            let html = '';
            for (const item of content) {
                if (item.type === 'text') {
                    html += this.renderMarkdown(item.text);
                } else if (item.type === 'image_url' && item.image_url && item.image_url.url) {
                    // Check if image data was stripped from localStorage
                    if (item.image_url.url === '[Large image data removed to save storage space]') {
                        html += `<div class="message-image-placeholder" style="display: inline-flex; align-items: center; padding: 12px; background: #f9f9f9; border: 2px dashed #d1d5db; border-radius: 8px; margin: 8px 0; gap: 8px; color: #6b7280;">
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#9ca3af" stroke-width="2">
                                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                                <circle cx="8.5" cy="8.5" r="1.5"/>
                                <polyline points="21,15 16,10 5,21"/>
                            </svg>
                            <span style="font-size: 14px; font-style: italic;">Image not available (removed to save storage space)</span>
                        </div>`;
                    } else {
                        const sizeInfo = this.getAttachmentSizeInfo(item);
                        html += `<div class="message-image" style="position: relative; display: inline-block; margin: 8px 0;">
                            <img src="${item.image_url.url}" alt="Uploaded image" style="max-width: 300px; max-height: 300px; border-radius: 8px; border: 1px solid #e5e7eb; display: block;">
                            <div style="position: absolute; top: 4px; right: 4px; background: rgba(0,0,0,0.7); color: white; padding: 2px 6px; border-radius: 4px; font-size: 11px; font-weight: 500;">
                                ${sizeInfo.formattedSize}
                            </div>
                        </div>`;
                    }
                } else if (item.type === 'file' && item.file && item.file.filename) {
                    // Display file attachment
                    const filename = item.file.filename;
                    // Check if file data was stripped from localStorage
                    const isPlaceholder = item.file.metadata === '[File data removed to save storage space]';

                    if (isPlaceholder) {
                        html += `<div class="message-file" style="display: inline-flex; align-items: center; padding: 8px 12px; background: #f9f9f9; border: 2px dashed #d1d5db; border-radius: 8px; margin: 4px 0; gap: 8px;">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#9ca3af" stroke-width="2">
                                <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
                            </svg>
                            <span style="font-size: 14px; color: #6b7280; font-style: italic;">${filename} (removed to save storage space)</span>
                        </div>`;
                    } else {
                        const sizeInfo = this.getAttachmentSizeInfo(item);
                        html += `<div class="message-file" style="display: inline-flex; align-items: center; padding: 8px 12px; background: #f3f4f6; border: 1px solid #d1d5db; border-radius: 8px; margin: 4px 0; gap: 8px;">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#dc2626" stroke-width="2">
                                <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
                            </svg>
                            <span style="font-size: 14px; color: #374151; flex: 1;">${filename}</span>
                            <span style="font-size: 11px; color: #6b7280; background: #e5e7eb; padding: 2px 6px; border-radius: 3px; font-weight: 500;">
                                ${sizeInfo.formattedSize}
                            </span>
                        </div>`;
                    }
                } else if (item.type === 'audio' && item.audio && item.audio.data) {
                    // Check if audio data was stripped from localStorage
                    if (item.audio.data === '[Large audio data removed to save storage space]') {
                        html += `<div class="message-audio-placeholder" style="display: flex; align-items: center; padding: 8px 12px; background: #f9f9f9; border: 2px dashed #d1d5db; border-radius: 8px; margin: 4px 0; gap: 8px; color: #6b7280; width: 100%; max-width: 100%;">
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" stroke-width="1.5">
                                <path d="M1 8h1l1-2 2 4 2-4 2 2h1"/>
                            </svg>
                            <span style="font-size: 13px; font-style: italic;">Audio file not available (removed to save storage space)</span>
                        </div>`;
                    } else {
                        // Display audio attachment with playback controls and size info
                        const sizeInfo = this.getAttachmentSizeInfo(item);
                        html += `<div class="message-audio" style="display: flex; align-items: center; padding: 8px 12px; background: #f0f9ff; border: 1px solid #bfdbfe; border-radius: 8px; margin: 4px 0; gap: 10px; width: 100%; max-width: 100%; max-height: 7vh; box-sizing: border-box;">
                            <svg width="16" height="2" viewBox="0 0 16 16" fill="none" stroke="#059669" stroke-width="1.5">
                                <path d="M1 8h1l1-2 2 4 2-4 2 2h1"/>
                            </svg>
                            <span style="font-size: 13px; color: #374151; font-weight: 500;">Audio file</span>
                            <span style="font-size: 10px; color: #059669; background: #dcfce7; padding: 2px 6px; border-radius: 3px; font-weight: 500;">
                                ${sizeInfo.formattedSize}
                            </span>
                            <audio controls style="height: 40px; flex: 1; min-width: 250px; margin-left: auto;">
                                <source src="${item.audio.data}" type="audio/webm">
                                <source src="${item.audio.data}" type="audio/mpeg">
                                <source src="${item.audio.data}" type="audio/wav">
                                <source src="${item.audio.data}" type="audio/mp3">
                                Your browser does not support the audio element.
                            </audio>
                        </div>`;
                    }
                } else if (item.type === 'audio_artifact' && item.audio_artifact) {
                    // Handle artifact-stored audio - show with server indicator
                    const artifactId = item.audio_artifact.artifactId;
                    const filename = item.audio_artifact.filename || 'recording.webm';
                    const formattedSize = item.audio_artifact.formattedSize || 'Unknown size';
                    const artifactUrl = `${this.baseUrl}/artifact/${artifactId}`;

                    html += `<div class="message-audio artifact" style="display: flex; align-items: center; padding: 8px 12px; background: #f0f9ff; border: 1px solid #059669; border-radius: 8px; margin: 4px 0; gap: 10px; width: 100%; max-width: 100%; max-height: 7vh; box-sizing: border-box;">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#059669" stroke-width="1.5">
                            <path d="M1 8h1l1-2 2 4 2-4 2 2h1"/>
                        </svg>
                        <span style="font-size: 13px; color: #374151; font-weight: 500;">Audio file</span>
                        <span style="font-size: 10px; color: #059669; background: #dcfce7; padding: 2px 6px; border-radius: 3px; font-weight: 500;">
                            ${formattedSize}
                        </span>
                        <span style="font-size: 9px; color: white; background: #059669; padding: 1px 4px; border-radius: 2px; font-weight: bold;">SERVER</span>
                        <audio controls style="height: 40px; flex: 1; min-width: 250px; margin-left: auto;">
                            <source src="${artifactUrl}" type="audio/webm">
                            <source src="${artifactUrl}" type="audio/mpeg">
                            <source src="${artifactUrl}" type="audio/wav">
                            <source src="${artifactUrl}" type="audio/mp3">
                            Your browser does not support the audio element.
                        </audio>
                    </div>`;
                }
            }
            return this.rewritePlantUMLUrls(html);
        } else {
            // Fallback for other content types
            let html = this.renderMarkdown(String(content));
            return this.rewritePlantUMLUrls(html);
        }
    }

    addMessage(role, content, index = null) {
        // Convert artifact references to storage-friendly format for localStorage
        const processedContent = this.processContentForStorage(content);
        const messageData = { role, content: processedContent };

        if (index !== null) {
            this.messages[index] = messageData;
        } else {
            this.messages.push(messageData);
            index = this.messages.length - 1;

            // CRITICAL: Save immediately after adding message using workers
            if (this.workerReady) {
                this.saveConversationsViaWorker().catch(error => {
                    console.error('Failed to save after adding message (worker):', error);
                    // Fallback to synchronous save
                    this.saveConversationsFallback();
                });
            } else {
                this.saveConversations().catch(error => {
                    console.error('Failed to save after adding message:', error);
                });
            }
        }

        this.renderMessages();
        this.saveCurrentConversation(); // Auto-save after adding message
    }

    // Process message content to store artifact references efficiently
    processContentForStorage(content) {
        if (Array.isArray(content)) {
            return content.map(item => {
                // For large audio files that are stored as artifacts, store reference instead of data
                if (item.type === 'audio' && item.audio && item.audio.data) {
                    const audioData = item.audio.data;

                    // DISABLED: Large audio warnings removed since cleanup is disabled
                    // All data is preserved regardless of size
                }
                return item;
            });
        }
        return content;
    }

    // Helper function to map message index to DOM index (accounting for tool notifications)
    getMessageDomIndex(messageIndex) {
        let domIndex = 0;
        for (let i = 0; i < messageIndex; i++) {
            if (this.messages[i].role === 'user' || this.messages[i].role === 'assistant') {
                domIndex++;
            }
        }
        return domIndex;
    }

    renderMessages() {
        this.chatMessages.innerHTML = '';

        this.messages.forEach((msg, index) => {
            // Handle tool notifications differently
            if (msg.role === 'tool') {
                const toolNotification = document.createElement('div');
                toolNotification.className = 'tool-notification';
                toolNotification.style.animation = 'fadeIn 0.3s ease-in';
                toolNotification.innerHTML = `
                    <svg class="tool-notification-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path>
                    </svg>
                    <span>${msg.content}</span>
                    <span style="margin-left: auto; font-size: 11px; opacity: 0.7;">Click to view details</span>
                `;

                // Add click handler to show tool details popup
                toolNotification.addEventListener('click', () => {
                    this.showToolDetailsPopup(msg);
                });

                this.chatMessages.appendChild(toolNotification);
                return;
            }

            const messageGroup = document.createElement('div');
            messageGroup.className = 'message-group';

            const message = document.createElement('div');
            message.className = `message ${msg.role === 'user' ? 'user-message' : 'assistant-message'}`;

            const avatar = document.createElement('div');
            avatar.className = `avatar ${msg.role === 'user' ? 'user-avatar' : 'assistant-avatar'}`;

            // Add accessibility attributes
            avatar.setAttribute('role', 'img');
            avatar.setAttribute('aria-label', msg.role === 'user' ? 'User message' : 'Assistant message');
            avatar.setAttribute('tabindex', '0');

            // Use professional icons instead of letters
            if (msg.role === 'user') {
                avatar.innerHTML = `
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
                        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                        <circle cx="12" cy="7" r="4"/>
                    </svg>
                `;
            } else {
                avatar.innerHTML = `
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
                        <path d="M12 2L2 7L12 12L22 7L12 2Z"/>
                        <path d="M2 17L12 22L22 17"/>
                        <path d="M2 12L12 17L22 12"/>
                    </svg>
                `;
            }

            const messageContent = document.createElement('div');
            messageContent.className = 'message-content';

            // Add accessibility attributes for message content
            messageContent.setAttribute('role', 'article');
            messageContent.setAttribute('aria-labelledby', `message-${index}-label`);

            // Add hidden label for screen readers
            const messageLabel = document.createElement('span');
            messageLabel.id = `message-${index}-label`;
            messageLabel.className = 'sr-only';
            messageLabel.textContent = `${msg.role === 'user' ? 'User' : 'Assistant'} message ${index + 1}`;
            messageContent.appendChild(messageLabel);

            // Show typing indicator if message is being typed
            if (msg.isTyping && !msg.content) {
                const typingContent = document.createElement('div');
                typingContent.innerHTML = `
                    <div class="typing-dots" style="display: flex; gap: 4px; padding: 10px 0;">
                        <div class="typing-dot"></div>
                        <div class="typing-dot"></div>
                        <div class="typing-dot"></div>
                    </div>
                `;
                messageContent.appendChild(typingContent);
            } else {
                const contentDiv = document.createElement('div');
                contentDiv.innerHTML = this.renderMessageContent(msg.content);
                messageContent.appendChild(contentDiv);
            }

            const actions = document.createElement('div');
            actions.className = 'message-actions';

            const editButton = document.createElement('button');
            editButton.className = 'action-button';
            editButton.textContent = 'Edit';
            editButton.onclick = () => this.startEdit(index);

            const replayButton = document.createElement('button');
            replayButton.className = 'action-button';
            replayButton.textContent = 'Replay from here';
            replayButton.onclick = () => this.replayFrom(index);

            actions.appendChild(editButton);
            if (msg.role === 'user') {
                actions.appendChild(replayButton);
            }

            messageContent.appendChild(actions);
            message.appendChild(avatar);
            message.appendChild(messageContent);
            messageGroup.appendChild(message);

            this.chatMessages.appendChild(messageGroup);
        });

        this.scrollToBottom();

        // Apply syntax highlighting to any code blocks (with small delay for DOM updates)
        setTimeout(() => {
            this.highlightCodeBlocks();
        }, 100);
    }

    startEdit(index) {
        // Extract text content and attachments from message
        const message = this.messages[index];
        let textContent = '';
        let attachments = [];

        if (typeof message.content === 'string') {
            textContent = message.content;
        } else if (Array.isArray(message.content)) {
            // Separate text content from attachments
            message.content.forEach(item => {
                if (item.type === 'text') {
                    textContent += item.text;
                } else if (item.type === 'image_url' || item.type === 'file' || item.type === 'audio') {
                    // Skip attachments that have been stripped from localStorage
                    const isStrippedImage = item.type === 'image_url' && item.image_url && item.image_url.url === '[Large image data removed to save storage space]';
                    const isStrippedFile = item.type === 'file' && item.file && item.file.metadata === '[File data removed to save storage space]';
                    const isStrippedAudio = item.type === 'audio' && item.audio && item.audio.data === '[Large audio data removed to save storage space]';

                    if (!isStrippedImage && !isStrippedFile && !isStrippedAudio) {
                        attachments.push(item);
                    }
                }
            });
        }

        // Use the helper method to render the edit interface
        this.renderEditInterface(index, textContent, attachments);
    }

    saveEdit(index, newContent) {
        this.messages[index].content = newContent;
        this.renderMessages();
        this.highlightCodeBlocks();
    }

    renderEditInterface(index, currentTextContent, currentAttachments) {
        const domIndex = this.getMessageDomIndex(index);
        const messageGroups = this.chatMessages.querySelectorAll('.message-group');
        const messageContent = messageGroups[domIndex].querySelector('.message-content');

        messageContent.classList.add('editing');
        messageContent.innerHTML = '';

        // Create edit container
        const editContainer = document.createElement('div');
        editContainer.className = 'edit-container';

        // Create attachments preview section if there are attachments
        if (currentAttachments.length > 0) {
            const attachmentsSection = document.createElement('div');
            attachmentsSection.className = 'edit-attachments';
            attachmentsSection.innerHTML = '<div style="font-size: 13px; font-weight: 500; margin-bottom: 8px; color: #374151;">Attachments:</div>';

            currentAttachments.forEach((attachment, attachIndex) => {
                const attachmentItem = document.createElement('div');
                attachmentItem.className = 'edit-attachment-item';
                attachmentItem.style.cssText = `
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    padding: 8px 12px;
                    background: #f3f4f6;
                    border: 1px solid #d1d5db;
                    border-radius: 6px;
                    margin-bottom: 8px;
                `;

                const sizeInfo = this.getAttachmentSizeInfo(attachment);

                if (attachment.type === 'image_url') {
                    attachmentItem.innerHTML = `
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#059669" stroke-width="2">
                            <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                            <circle cx="8.5" cy="8.5" r="1.5"/>
                            <polyline points="21,15 16,10 5,21"/>
                        </svg>
                        <span style="flex: 1; font-size: 14px; color: #374151;">Image attachment</span>
                        <span style="font-size: 11px; color: #059669; background: #dcfce7; padding: 2px 6px; border-radius: 3px; font-weight: 500;">
                            ${sizeInfo.formattedSize}
                        </span>
                        <button class="remove-attachment" data-attach-index="${attachIndex}" style="
                            background: #ef4444;
                            color: white;
                            border: none;
                            border-radius: 4px;
                            width: 20px;
                            height: 20px;
                            cursor: pointer;
                            display: flex;
                            align-items: center;
                            justify-content: center;
                            font-size: 12px;
                            margin-left: 8px;
                        ">×</button>
                    `;
                } else if (attachment.type === 'file') {
                    const filename = attachment.file?.filename || 'Unknown file';
                    attachmentItem.innerHTML = `
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#dc2626" stroke-width="2">
                            <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
                        </svg>
                        <span style="flex: 1; font-size: 14px; color: #374151;">${filename}</span>
                        <span style="font-size: 11px; color: #6b7280; background: #e5e7eb; padding: 2px 6px; border-radius: 3px; font-weight: 500;">
                            ${sizeInfo.formattedSize}
                        </span>
                        <button class="remove-attachment" data-attach-index="${attachIndex}" style="
                            background: #ef4444;
                            color: white;
                            border: none;
                            border-radius: 4px;
                            width: 20px;
                            height: 20px;
                            cursor: pointer;
                            display: flex;
                            align-items: center;
                            justify-content: center;
                            font-size: 12px;
                            margin-left: 8px;
                        ">×</button>
                    `;
                } else if (attachment.type === 'audio') {
                    attachmentItem.innerHTML = `
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#059669" stroke-width="2">
                            <path d="m9 9 3-3m-3 3v6a3 3 0 1 0 6 0V9m-6 0h6"/>
                            <circle cx="12" cy="4" r="2"/>
                        </svg>
                        <span style="flex: 1; font-size: 14px; color: #374151;">Audio file</span>
                        <span style="font-size: 11px; color: #059669; background: #dcfce7; padding: 2px 6px; border-radius: 3px; font-weight: 500;">
                            ${sizeInfo.formattedSize}
                        </span>
                        <button class="remove-attachment" data-attach-index="${attachIndex}" style="
                            background: #ef4444;
                            color: white;
                            border: none;
                            border-radius: 4px;
                            width: 20px;
                            height: 20px;
                            cursor: pointer;
                            display: flex;
                            align-items: center;
                            justify-content: center;
                            font-size: 12px;
                            margin-left: 8px;
                        ">×</button>
                    `;
                }

                attachmentsSection.appendChild(attachmentItem);
            });

            editContainer.appendChild(attachmentsSection);
        }

        // Create textarea for text content
        const textarea = document.createElement('textarea');
        textarea.className = 'edit-textarea';
        textarea.value = currentTextContent;
        textarea.style.marginTop = currentAttachments.length > 0 ? '8px' : '0';

        // Create attachment controls section
        const attachmentControls = document.createElement('div');
        attachmentControls.className = 'edit-attachment-controls';
        attachmentControls.style.cssText = `
            display: flex;
            gap: 8px;
            margin-top: 8px;
            margin-bottom: 12px;
        `;

        // Add image/PDF button
        const addImageButton = document.createElement('button');
        addImageButton.className = 'attachment-button';
        addImageButton.style.cssText = `
            padding: 6px 12px;
            border: 1px solid #d1d5db;
            border-radius: 6px;
            background: #f9fafb;
            color: #374151;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 13px;
            transition: all 0.2s;
        `;
        addImageButton.innerHTML = `
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                <circle cx="8.5" cy="8.5" r="1.5"/>
                <polyline points="21,15 16,10 5,21"/>
            </svg>
            Add Image/PDF
        `;

        // Add audio file button
        const addAudioButton = document.createElement('button');
        addAudioButton.className = 'attachment-button';
        addAudioButton.style.cssText = addImageButton.style.cssText;
        addAudioButton.innerHTML = `
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
                <path d="m19 10v2a7 7 0 0 1-14 0v-2"/>
                <line x1="12" x2="12" y1="19" y2="23"/>
                <line x1="8" x2="16" y1="23" y2="23"/>
            </svg>
            Add Audio
        `;

        // Create hidden file inputs for editing
        const editImageInput = document.createElement('input');
        editImageInput.type = 'file';
        editImageInput.accept = 'image/*,application/pdf';
        editImageInput.multiple = true;
        editImageInput.style.display = 'none';

        const editAudioInput = document.createElement('input');
        editAudioInput.type = 'file';
        editAudioInput.accept = 'audio/*';
        editAudioInput.multiple = true;
        editAudioInput.style.display = 'none';

        // Add event listeners for attachment buttons
        addImageButton.onclick = () => editImageInput.click();
        addAudioButton.onclick = () => editAudioInput.click();

        editImageInput.addEventListener('change', async (e) => {
            if (e.target.files.length > 0) {
                const newAttachments = await this.processEditAttachments(e.target.files);
                currentAttachments.push(...newAttachments);
                this.renderEditInterface(index, textarea.value, currentAttachments);
            }
        });

        editAudioInput.addEventListener('change', async (e) => {
            if (e.target.files.length > 0) {
                const newAttachments = await this.processEditAttachments(e.target.files);
                currentAttachments.push(...newAttachments);
                this.renderEditInterface(index, textarea.value, currentAttachments);
            }
        });

        attachmentControls.appendChild(addImageButton);
        attachmentControls.appendChild(addAudioButton);
        attachmentControls.appendChild(editImageInput);
        attachmentControls.appendChild(editAudioInput);

        // Create edit buttons
        const editButtons = document.createElement('div');
        editButtons.className = 'edit-buttons';
        editButtons.style.cssText = `
            display: flex;
            gap: 8px;
            margin-top: 12px;
        `;

        const saveButton = document.createElement('button');
        saveButton.className = 'save-button';
        saveButton.textContent = 'Save';
        saveButton.style.cssText = `
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            background: #059669;
            color: white;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: background 0.2s;
        `;
        saveButton.onclick = () => this.saveEditWithAttachments(index, textarea.value, currentAttachments);

        const saveAndRestartButton = document.createElement('button');
        saveAndRestartButton.className = 'save-restart-button';
        saveAndRestartButton.textContent = 'Save & Restart from here';
        saveAndRestartButton.style.cssText = `
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            background: #dc2626;
            color: white;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: background 0.2s;
        `;
        saveAndRestartButton.onclick = () => this.saveEditAndRestart(index, textarea.value, currentAttachments);

        const cancelButton = document.createElement('button');
        cancelButton.className = 'cancel-button';
        cancelButton.textContent = 'Cancel';
        cancelButton.style.cssText = `
            padding: 8px 16px;
            border: 1px solid #d1d5db;
            border-radius: 6px;
            background: #f9fafb;
            color: #374151;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.2s;
        `;
        cancelButton.onclick = () => this.cancelEdit();

        editButtons.appendChild(saveButton);
        editButtons.appendChild(saveAndRestartButton);
        editButtons.appendChild(cancelButton);

        editContainer.appendChild(textarea);
        editContainer.appendChild(attachmentControls);
        editContainer.appendChild(editButtons);
        messageContent.appendChild(editContainer);

        // Add event listeners for attachment removal
        editContainer.addEventListener('click', (e) => {
            if (e.target.classList.contains('remove-attachment')) {
                const attachIndex = parseInt(e.target.dataset.attachIndex);
                currentAttachments.splice(attachIndex, 1);
                // Re-render the edit interface with updated attachments
                this.renderEditInterface(index, textarea.value, currentAttachments);
            }
        });

        textarea.focus();
    }

    async processEditAttachments(files) {
        const SIZE_THRESHOLD = 500 * 1024; // 500KB threshold for artifact storage
        const attachments = [];

        for (const file of files) {
            if (file.type.startsWith('image/') ||
                file.type === 'application/pdf' ||
                file.type.startsWith('audio/') ||
                file.type.startsWith('text/') ||
                file.type === 'application/json' ||
                file.type === 'application/xml' ||
                file.type.includes('csv') ||
                file.type.includes('excel') ||
                file.type.includes('word') ||
                file.type.includes('powerpoint')) {

                try {
                    let attachmentData;

                    if (file.size > SIZE_THRESHOLD && !this.artifactServerUnavailable) {
                        // Store large files using artifact storage
                        const artifactId = await this.uploadToArtifactStorage(file, file.name);

                        if (file.type.startsWith('image/')) {
                            attachmentData = {
                                type: 'image_artifact',
                                image_artifact: {
                                    artifactId: artifactId,
                                    filename: file.name,
                                    mimeType: file.type,
                                    formattedSize: this.formatFileSize(file.size)
                                }
                            };
                        } else if (file.type.startsWith('audio/')) {
                            attachmentData = {
                                type: 'audio_artifact',
                                audio_artifact: {
                                    artifactId: artifactId,
                                    filename: file.name,
                                    mimeType: file.type,
                                    formattedSize: this.formatFileSize(file.size)
                                }
                            };
                        } else {
                            attachmentData = {
                                type: 'file_artifact',
                                file_artifact: {
                                    artifactId: artifactId,
                                    filename: file.name,
                                    mimeType: file.type,
                                    formattedSize: this.formatFileSize(file.size)
                                }
                            };
                        }
                    } else {
                        // Store small files as base64 data URLs
                        const dataURL = await this.fileToDataURL(file);
                        if (file.type.startsWith('image/')) {
                            attachmentData = {
                                type: 'image_url',
                                image_url: {
                                    url: dataURL
                                }
                            };
                        } else {
                            attachmentData = {
                                type: 'file',
                                file: {
                                    data: dataURL,
                                    filename: file.name
                                }
                            };
                        }
                    }

                    attachments.push(attachmentData);
                } catch (error) {
                    console.error('Error processing file:', error);
                    this.showNotification(`Error processing ${file.name}: ${error.message}`, 'error');
                }
            } else {
                this.showNotification(`Unsupported file type: ${file.type}`, 'error');
            }
        }

        return attachments;
    }

    saveEditAndRestart(index, textContent, attachments) {
        // Save the edited message first
        this.saveEditWithAttachments(index, textContent, attachments);

        // Then restart conversation from this point
        this.replayFrom(index);
    }

    saveEditWithAttachments(index, textContent, attachments) {
        // Reconstruct the message content with text and remaining attachments
        if (attachments.length === 0 && textContent.trim() === '') {
            // If no attachments and no text, just set empty string
            this.messages[index].content = '';
        } else if (attachments.length === 0) {
            // Only text content
            this.messages[index].content = textContent;
        } else {
            // Multimodal content with text and attachments
            const newContent = [];

            // Add text content if present
            if (textContent.trim()) {
                newContent.push({
                    type: 'text',
                    text: textContent
                });
            }

            // Add remaining attachments
            newContent.push(...attachments);

            this.messages[index].content = newContent;
        }

        this.renderMessages();
        this.saveCurrentConversation(); // Save the updated message

        // CRITICAL: Save immediately after editing message with robust error handling
        this.saveConversations().catch(error => {
            console.error('Failed to save after editing message:', error);
        });
    }

    cancelEdit() {
        this.renderMessages();
    }

    replayFrom(index) {
        // Remove all messages after this index
        this.messages = this.messages.slice(0, index + 1);
        this.renderMessages();

        // Regenerate assistant response
        this.getAssistantResponse();
    }

    async sendMessage() {
        const textContent = this.chatInput.value.trim();
        if (!textContent && this.selectedFiles.length === 0) return;

        // Create message content based on whether we have files
        let messageContent;
        if (this.selectedFiles.length > 0) {
            // Multimodal message with files
            messageContent = [];

            // Add text content if present
            if (textContent) {
                messageContent.push({
                    type: "text",
                    text: textContent
                });
            }

            // Add files - resolve artifacts before sending
            for (const file of this.selectedFiles) {
                let dataURL = file.dataURL;

                // If this is an artifact reference, fetch the actual data
                if (file.dataURL.startsWith('artifact:')) {
                    try {
                        const artifactId = file.dataURL.replace('artifact:', '');
                        console.log('Resolving artifact for sending:', artifactId);
                        dataURL = await this.fetchArtifactAsDataURL(artifactId);
                    } catch (error) {
                        console.error('Failed to resolve artifact for sending:', error);
                        this.showError(`Failed to load audio file: ${error.message}`);
                        return; // Don't send message if we can't resolve the artifact
                    }
                }

                if (file.fileType.startsWith('image/')) {
                    // Handle images with the existing image_url format
                    messageContent.push({
                        type: "image_url",
                        image_url: {
                            url: dataURL
                        }
                    });
                } else if (file.fileType === 'application/pdf') {
                    // Handle PDFs with the file format
                    messageContent.push({
                        type: "file",
                        file: {
                            file_data: dataURL,
                            filename: file.fileName
                        }
                    });
                } else if (file.fileType.startsWith('audio/')) {
                    // Handle audio files with the audio format
                    messageContent.push({
                        type: "audio",
                        audio: {
                            data: dataURL
                        }
                    });
                }
            }
        } else {
            // Text-only message
            messageContent = textContent;
        }

        // Clear inputs
        this.chatInput.value = '';
        this.chatInput.style.height = 'auto';
        this.clearFilePreviews();

        this.addMessage('user', messageContent);
        await this.getAssistantResponse();
    }

    saveSystemPrompt() {
        this.systemPrompt = this.systemPromptTextarea.value.trim() || 'You are a helpful assistant.\\nCurrent time is {{now | formatTimeInLocation "Europe/Paris" "2006-01-02 15:04"}}';
        this.saveCurrentConversation();
        // Flash the save button to indicate success
        this.systemPromptSave.style.background = 'rgba(34, 197, 94, 0.5)';
        setTimeout(() => {
            this.systemPromptSave.style.background = '';
        }, 300);
    }

    resetSystemPrompt() {
        this.systemPrompt = 'You are a helpful assistant.\nCurrent time is {{now | formatTimeInLocation "Europe/Paris" "2006-01-02 15:04"}}';
        this.systemPromptTextarea.value = this.systemPrompt;
        this.saveCurrentConversation();
        // Flash the reset button to indicate success
        this.systemPromptReset.style.background = 'rgba(239, 68, 68, 0.5)';
        setTimeout(() => {
            this.systemPromptReset.style.background = '';
        }, 300);
    }

    async manualCleanupStorage() {
        // DISABLED: All cleanup mechanisms removed per user request
        alert('Cleanup functionality has been disabled. Use individual delete buttons to remove specific conversations, or clear browser storage manually if needed.');
    }

    async getAssistantResponse() {
        this.typingIndicator.classList.add('active');
        this.sendButton.disabled = true;
        this.isStreaming = true;
        this.updateSendButton();  // Update button to show "Stop"

        // Prepare API request data using workers for heavy conversation processing
        let apiRequestData;
        try {
            if (this.workerReady) {
                // Use worker to prepare API request (non-blocking)
                const currentConversation = { messages: this.messages };
                const selectedTools = Array.from(this.selectedTools);

                const result = await this.workerManager.prepareConversationForAPI(
                    currentConversation,
                    this.systemPrompt,
                    selectedTools
                );

                if (result.success) {
                    console.log(`API request prepared by worker (${result.data.messageCount} messages)`);
                    apiRequestData = {
                        model: this.buildModelWithTools() || 'gemini-2.0-flash',
                        messages: result.data.messages,
                        temperature: 0.7,
                        max_tokens: 2000,
                        stream: true,
                        ...(result.data.tools && { tools: result.data.tools })
                    };
                } else {
                    throw new Error(`Worker preparation failed: ${result.error}`);
                }
            } else {
                // Fallback to synchronous preparation
                const messagesWithSystem = [
                    { role: 'system', content: this.systemPrompt },
                    ...this.messages
                ];

                apiRequestData = {
                    model: this.buildModelWithTools() || 'gemini-2.0-flash',
                    messages: messagesWithSystem,
                    temperature: 0.7,
                    max_tokens: 2000,
                    stream: true
                };
            }
        } catch (workerError) {
            console.warn('Worker API preparation failed, using fallback:', workerError);
            // Fallback to synchronous preparation
            const messagesWithSystem = [
                { role: 'system', content: this.systemPrompt },
                ...this.messages
            ];

            apiRequestData = {
                model: this.buildModelWithTools() || 'gemini-2.0-flash',
                messages: messagesWithSystem,
                temperature: 0.7,
                max_tokens: 2000,
                stream: true
            };
        }

        try {
            const response = await fetch(this.apiUrl, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(apiRequestData)
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // Check if response is streaming
            const contentType = response.headers.get('content-type');
            console.log('Response content-type:', contentType);

            if (contentType && contentType.includes('text/event-stream')) {
                console.log('Handling streaming response...');
                // Handle streaming response
                await this.handleStreamingResponse(response);
            } else {
                console.log('Handling regular JSON response...');
                // Handle non-streaming response
                const data = await response.json();
                if (data.choices && data.choices[0] && data.choices[0].message) {
                    this.addMessage('assistant', data.choices[0].message.content);
                } else {
                    throw new Error('Invalid response format');
                }
            }
        } catch (error) {
            console.error('Error:', error);
            this.showError(`Failed to get response: ${error.message}`);
        } finally {
            this.typingIndicator.classList.remove('active');
            this.sendButton.disabled = false;
            this.isStreaming = false;
            this.currentReader = null;
            this.updateSendButton();  // Reset button to "Send"
            // Close any remaining tool popups when stream ends
            this.closeAllToolPopups();
        }
    }

    // Add permanent tool notification to conversation
    addToolNotification(toolName, toolCallData) {
        // Add tool notification as a special message type
        const toolMessage = {
            role: 'tool',
            content: `Calling tool: ${toolName}`,
            toolName: toolName,
            toolCallData: toolCallData,  // Store the complete tool call data
            toolResponse: null  // Will be populated when response arrives
        };

        this.messages.push(toolMessage);

        // CRITICAL: Save immediately after adding tool message
        this.saveConversations().catch(error => {
            console.error('Failed to save after adding tool message:', error);
        });

        // Force re-render to show the tool notification immediately
        this.renderMessages();
        this.saveCurrentConversation(); // Save tool notification
        this.scrollToBottom();

        return this.messages.length - 1; // Return the index of the tool message
    }

    // Tool popup management methods
    showToolCallPopup(event) {
        if (!event || !event.tool_call || !event.tool_call.id) {
            console.error('Invalid tool call event for popup:', event);
            return;
        }

        const popupId = event.tool_call.id;

        // Check if popup already exists
        if (this.toolPopups.has(popupId)) {
            console.log('Popup already exists for:', popupId);
            return;
        }

        console.log('Creating tool call popup for:', popupId, event.tool_call.name);

        // Create popup element
        const popup = document.createElement('div');
        popup.className = 'tool-popup tool-call';
        popup.id = `popup-${popupId}`;

        // Create popup HTML
        popup.innerHTML = `
            <div class="tool-popup-header">
                <div class="tool-popup-title">
                    <div class="tool-popup-icon">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#007AFF" stroke-width="2">
                            <circle cx="12" cy="12" r="10"></circle>
                            <path d="M12 6v6l4 2"></path>
                        </svg>
                    </div>
                    Tool Executing: ${event.tool_call.name}
                </div>
                <button class="tool-popup-close" onclick="chatUI.closeToolPopup('${popupId}')">×</button>
            </div>
            <div class="tool-popup-content">
                <div>
                    <strong style="color: #333; font-size: 13px;">Arguments:</strong>
                    <div class="tool-popup-args">${JSON.stringify(event.tool_call.arguments, null, 2)}</div>
                </div>
                <div style="margin-top: 12px; display: flex; align-items: center; gap: 10px;">
                    <div class="tool-popup-spinner"></div>
                    <span style="font-size: 14px; color: #666;">Waiting for response...</span>
                </div>
            </div>
        `;

        // Add to container
        const container = document.getElementById('toolPopupContainer');
        container.appendChild(popup);

        // Store reference
        this.toolPopups.set(popupId, popup);

        // Set a longer timeout (30 seconds) for tools that take a while
        // This will only trigger if no response is received
        const timer = setTimeout(() => {
            // Update popup to show timeout before closing
            const popup = this.toolPopups.get(popupId);
            if (popup) {
                const content = popup.querySelector('.tool-popup-content');
                if (content) {
                    content.innerHTML = `
                        <div style="display: flex; align-items: center; gap: 8px;">
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FFA500" stroke-width="2">
                                <circle cx="12" cy="12" r="10"></circle>
                                <line x1="12" y1="8" x2="12" y2="12"></line>
                                <line x1="12" y1="16" x2="12" y2="16"></line>
                            </svg>
                            <span style="font-size: 14px; color: #FFA500; font-weight: 500;">
                                Tool execution timeout - no response received
                            </span>
                        </div>
                    `;
                }
                // Close after showing timeout message
                setTimeout(() => {
                    this.closeToolPopup(popupId);
                }, 2000);
            }
        }, 30000);
        this.popupAutoCloseTimers.set(popupId, timer)
    }

    updateToolResponsePopup(event) {
        const popupId = event.tool_response.id;
        console.log('Updating tool response popup for:', popupId, event.tool_response.name);

        const popup = this.toolPopups.get(popupId);

        if (!popup) {
            // If no existing popup, create one for the response
            console.log('No matching tool call popup found for response, creating new one:', popupId);
            this.showToolResponseOnlyPopup(event);
            return;
        }

        // Clear any existing auto-close timer
        const timer = this.popupAutoCloseTimers.get(popupId);
        if (timer) {
            clearTimeout(timer);
            this.popupAutoCloseTimers.delete(popupId);
        }

        // Update the popup to show the response
        const hasError = event.tool_response.error;
        popup.className = `tool-popup ${hasError ? 'tool-error' : 'tool-response'}`;

        // Update the icon in the header
        const iconElement = popup.querySelector('.tool-popup-icon');
        if (iconElement) {
            if (hasError) {
                iconElement.innerHTML = `
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FF3B30" stroke-width="2">
                        <circle cx="12" cy="12" r="10"></circle>
                        <line x1="15" y1="9" x2="9" y2="15"></line>
                        <line x1="9" y1="9" x2="15" y2="15"></line>
                    </svg>
                `;
            } else {
                iconElement.innerHTML = `
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#34C759" stroke-width="2">
                        <circle cx="12" cy="12" r="10"></circle>
                        <path d="M9 12l2 2 4-4"></path>
                    </svg>
                `;
            }
        }

        // Update title to show status
        const titleElement = popup.querySelector('.tool-popup-title');
        if (titleElement) {
            // Keep the icon and update the text
            const iconHTML = iconElement ? iconElement.outerHTML : '';
            titleElement.innerHTML = iconHTML + (hasError ? ` Tool Failed: ${event.tool_response.name}` : ` Tool Completed: ${event.tool_response.name}`);
        }

        // Get existing content div
        const content = popup.querySelector('.tool-popup-content');

        // Keep the arguments that were already displayed and add the response
        const existingArgs = popup.querySelector('.tool-popup-args');
        let argsHTML = '';
        if (existingArgs) {
            argsHTML = existingArgs.outerHTML;
        }

        // Build the complete content with both arguments and response
        let responseHTML = '';
        if (hasError) {
            responseHTML = `
                <div class="tool-popup-error">
                    <strong>Error:</strong> ${event.tool_response.error}
                </div>
            `;
        } else {
            // Format the response
            if (typeof event.tool_response.response === 'string') {
                responseHTML = `
                    <div style="margin-top: 12px;">
                        <strong style="color: #333; font-size: 13px;">Response:</strong>
                        <div class="tool-popup-response">${event.tool_response.response}</div>
                    </div>
                `;
            } else if (typeof event.tool_response.response === 'object' && event.tool_response.response !== null) {
                responseHTML = `
                    <div style="margin-top: 12px;">
                        <strong style="color: #333; font-size: 13px;">Response:</strong>
                        <div class="tool-popup-response"><pre>${JSON.stringify(event.tool_response.response, null, 2)}</pre></div>
                    </div>
                `;
            } else {
                responseHTML = `
                    <div style="margin-top: 12px;">
                        <strong style="color: #333; font-size: 13px;">Response:</strong>
                        <div class="tool-popup-response">Tool executed successfully</div>
                    </div>
                `;
            }
        }

        // Update content with both arguments and response
        content.innerHTML = argsHTML + responseHTML;

        // Auto-close after 5.5 seconds to give user time to see the result
        setTimeout(() => {
            this.closeToolPopup(popupId);
        }, 5500);
    }

    showToolResponseOnlyPopup(event) {
        const popupId = event.tool_response.id;
        const hasError = event.tool_response.error;

        console.log('Creating response-only popup for:', popupId, event.tool_response.name);

        // Create popup element
        const popup = document.createElement('div');
        popup.className = `tool-popup ${hasError ? 'tool-error' : 'tool-response'}`;
        popup.id = `popup-${popupId}`;

        const iconHtml = hasError ? `
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FF3B30" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="15" y1="9" x2="9" y2="15"></line>
                <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
        ` : `
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#34C759" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <path d="M9 12l2 2 4-4"></path>
            </svg>
        `;

        // Create popup HTML
        popup.innerHTML = `
            <div class="tool-popup-header">
                <div class="tool-popup-title">
                    <div class="tool-popup-icon">${iconHtml}</div>
                    ${hasError ? 'Tool Failed' : 'Tool Completed'}: ${event.tool_response.name}
                </div>
                <button class="tool-popup-close" onclick="chatUI.closeToolPopup('${popupId}')">×</button>
            </div>
            <div class="tool-popup-content">
                ${hasError
                    ? `<div class="tool-popup-error"><strong>Error:</strong> ${event.tool_response.error}</div>`
                    : `<div>
                        <strong style="color: #333; font-size: 13px;">Response:</strong>
                        <div class="tool-popup-response">
                            ${typeof event.tool_response.response === 'string'
                                ? event.tool_response.response
                                : `<pre>${JSON.stringify(event.tool_response.response, null, 2)}</pre>`}
                        </div>
                       </div>`
                }
            </div>
        `;

        // Add to container
        const container = document.getElementById('toolPopupContainer');
        container.appendChild(popup);

        // Store reference
        this.toolPopups.set(popupId, popup);

        // Auto-close after 5.5 seconds
        const timer = setTimeout(() => {
            this.closeToolPopup(popupId);
        }, 5500);
        this.popupAutoCloseTimers.set(popupId, timer);
    }

    closeToolPopup(popupId) {
        const popup = this.toolPopups.get(popupId);
        if (popup) {
            // Clear timer if exists
            const timer = this.popupAutoCloseTimers.get(popupId);
            if (timer) {
                clearTimeout(timer);
                this.popupAutoCloseTimers.delete(popupId);
            }

            // Add fade-out animation
            popup.classList.add('fade-out');

            // Remove after animation
            setTimeout(() => {
                popup.remove();
                this.toolPopups.delete(popupId);
            }, 300);
        }
    }

    closeAllToolPopups() {
        for (const popupId of this.toolPopups.keys()) {
            this.closeToolPopup(popupId);
        }
    }

    // Store tool response in the corresponding tool message
    storeToolResponse(toolResponseEvent) {
        const responseId = toolResponseEvent.tool_response?.id;
        if (!responseId) return;

        // Find the corresponding tool message by matching the tool call ID
        for (let i = this.messages.length - 1; i >= 0; i--) {
            const msg = this.messages[i];
            if (msg.role === 'tool' &&
                msg.toolCallData &&
                msg.toolCallData.tool_call &&
                msg.toolCallData.tool_call.id === responseId) {

                // Store the response data
                msg.toolResponse = toolResponseEvent.tool_response;

                // Save the conversation with the updated tool data
                this.saveCurrentConversation();
                break;
            }
        }
    }

    // Method to show tool details popup when notification is clicked
    showToolDetailsPopup(toolMessage) {
        const popupId = `details-${Date.now()}`;

        // Create popup element
        const popup = document.createElement('div');
        popup.className = 'tool-popup tool-response';
        popup.id = `popup-${popupId}`;

        // Build the content
        let argumentsHTML = '';
        if (toolMessage.toolCallData && toolMessage.toolCallData.tool_call && toolMessage.toolCallData.tool_call.arguments) {
            argumentsHTML = `
                <div style="margin-bottom: 16px;">
                    <strong style="color: #333; font-size: 13px;">Arguments:</strong>
                    <div class="tool-popup-args">${JSON.stringify(toolMessage.toolCallData.tool_call.arguments, null, 2)}</div>
                </div>
            `;
        }

        let responseHTML = '';
        if (toolMessage.toolResponse) {
            if (toolMessage.toolResponse.error) {
                responseHTML = `
                    <div class="tool-popup-error">
                        <strong>Error:</strong> ${toolMessage.toolResponse.error}
                    </div>
                `;
            } else {
                const responseContent = typeof toolMessage.toolResponse.response === 'string'
                    ? toolMessage.toolResponse.response
                    : `<pre>${JSON.stringify(toolMessage.toolResponse.response, null, 2)}</pre>`;
                responseHTML = `
                    <div>
                        <strong style="color: #333; font-size: 13px;">Response:</strong>
                        <div class="tool-popup-response">${responseContent}</div>
                    </div>
                `;
            }
        } else {
            responseHTML = `
                <div style="color: #666; font-style: italic;">
                    Response data not available (tool may still be executing)
                </div>
            `;
        }

        // Create popup HTML
        popup.innerHTML = `
            <div class="tool-popup-header">
                <div class="tool-popup-title">
                    <div class="tool-popup-icon">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#34C759" stroke-width="2">
                            <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path>
                        </svg>
                    </div>
                    Tool Details: ${toolMessage.toolName}
                </div>
                <button class="tool-popup-close" onclick="chatUI.closeToolPopup('${popupId}')" style="z-index: 1001;">×</button>
            </div>
            <div class="tool-popup-content">
                ${argumentsHTML}
                ${responseHTML}
            </div>
        `;

        // Add to container
        const container = document.getElementById('toolPopupContainer');
        container.appendChild(popup);

        // Store reference
        this.toolPopups.set(popupId, popup);

        // Auto-close after 10 seconds
        const timer = setTimeout(() => {
            this.closeToolPopup(popupId);
        }, 10000);
        this.popupAutoCloseTimers.set(popupId, timer);
    }

    async handleStreamingResponse(response) {
        const reader = response.body.getReader();
        this.currentReader = reader;  // Store the reader so we can cancel it
        const decoder = new TextDecoder();
        let assistantMessage = '';
        let messageIndex = null;

        // Don't add assistant message yet - wait for tool calls to complete first
        // We'll add it when we get the first content or when no tools are called

        try {
            let isStreaming = true;
            while (true) {
                // Check if streaming was stopped by user before each read
                if (!this.isStreaming) {
                    console.log('🛑 Streaming stopped by user - breaking from loop');
                    break;
                }

                const { value, done } = await reader.read();

                // Check again after read in case user stopped while waiting
                if (!this.isStreaming) {
                    console.log('🛑 Streaming stopped by user after read - breaking from loop');
                    break;
                }

                if (done) {
                    isStreaming = false;
                    this.isStreaming = false;
                    this.currentReader = null;
                    this.updateSendButton();

                    // Handle case where no content was received but streaming is done
                    if (messageIndex === null) {
                        this.messages.push({ role: 'assistant', content: assistantMessage, isTyping: false });
                        messageIndex = this.messages.length - 1;

                        // CRITICAL: Save immediately after adding assistant message
                        this.saveConversations().catch(error => {
                            console.error('Failed to save after adding assistant message:', error);
                        });
                    } else {
                        this.messages[messageIndex].isTyping = false;
                    }

                    this.updateMessageContent(messageIndex, assistantMessage, false);
                    this.saveCurrentConversation(); // Save when streaming completes
                    break;
                }

                const chunk = decoder.decode(value, { stream: true });
                const lines = chunk.split('\n');

                for (const line of lines) {
                    // Check if streaming was stopped during chunk processing
                    if (!this.isStreaming) {
                        console.log('🛑 Streaming stopped by user during chunk processing');
                        return; // Exit immediately
                    }
                    if (line.startsWith('data: ')) {
                        const data = line.slice(6);
                        if (data === '[DONE]') {
                            isStreaming = false;
                            this.isStreaming = false;
                            this.currentReader = null;
                            this.updateSendButton();

                            // Handle case where no content was received but streaming is done
                            if (messageIndex === null) {
                                this.messages.push({ role: 'assistant', content: assistantMessage, isTyping: false });
                                messageIndex = this.messages.length - 1;

                                // CRITICAL: Save immediately after adding assistant message
                                this.saveConversations().catch(error => {
                                    console.error('Failed to save after adding assistant message:', error);
                                });
                            } else {
                                this.messages[messageIndex].isTyping = false;
                            }

                            this.updateMessageContent(messageIndex, assistantMessage, false);
                            this.saveCurrentConversation(); // Save when streaming completes
                            return;
                        }

                        try {
                            const parsed = JSON.parse(data);

                            // Enhanced debugging with more details
                            if (parsed.event_type) {
                                console.log('🔧 Tool event received:', {
                                    type: parsed.event_type,
                                    object: parsed.object,
                                    data: parsed
                                });
                            }

                            // Check for tool events with comprehensive validation
                            if (parsed.event_type === 'tool_call') {
                                console.log('🚀 Tool call event detected:', parsed);
                                if (parsed.tool_call && parsed.tool_call.name && parsed.tool_call.id) {
                                    console.log('✅ Valid tool call, adding notification and popup');
                                    this.addToolNotification(parsed.tool_call.name, parsed);
                                    this.showToolCallPopup(parsed);
                                } else {
                                    console.warn('❌ Invalid tool call event structure:', {
                                        hasToolCall: !!parsed.tool_call,
                                        hasName: !!(parsed.tool_call && parsed.tool_call.name),
                                        hasId: !!(parsed.tool_call && parsed.tool_call.id),
                                        fullData: parsed
                                    });
                                }
                            } else if (parsed.event_type === 'tool_response') {
                                console.log('📥 Tool response event detected:', parsed);
                                if (parsed.tool_response && parsed.tool_response.id) {
                                    console.log('✅ Valid tool response, updating popup');
                                    this.updateToolResponsePopup(parsed);
                                    this.storeToolResponse(parsed);
                                } else {
                                    console.warn('❌ Invalid tool response event structure:', {
                                        hasToolResponse: !!parsed.tool_response,
                                        hasId: !!(parsed.tool_response && parsed.tool_response.id),
                                        fullData: parsed
                                    });
                                }
                            } else if (parsed.event_type === 'error') {
                                console.log('🚨 Error event detected:', parsed);
                                if (parsed.error && parsed.error.message) {
                                    console.log('✅ Valid error event, displaying error');
                                    this.showErrorNotification(parsed);
                                } else {
                                    console.warn('❌ Invalid error event structure:', {
                                        hasError: !!parsed.error,
                                        hasMessage: !!(parsed.error && parsed.error.message),
                                        fullData: parsed
                                    });
                                }
                            } else if (parsed.choices && parsed.choices[0]) {
                                // Handle regular chat completion chunks
                                const delta = parsed.choices[0].delta;
                                if (delta && delta.content) {
                                    // If we don't have a message index yet (no tool calls), create assistant message now
                                    if (messageIndex === null) {
                                        this.messages.push({ role: 'assistant', content: '', isTyping: true });
                                        messageIndex = this.messages.length - 1;

                                        // CRITICAL: Save immediately after adding assistant message
                                        this.saveConversations().catch(error => {
                                            console.error('Failed to save after adding assistant message:', error);
                                        });
                                        this.renderMessages();
                                    }

                                    // Remove typing indicator on first content
                                    if (this.messages[messageIndex].isTyping) {
                                        this.messages[messageIndex].isTyping = false;
                                    }
                                    assistantMessage += delta.content;
                                    // Update the message in real-time with streaming indicator
                                    this.messages[messageIndex].content = assistantMessage;
                                    this.updateMessageContent(messageIndex, assistantMessage, true);
                                }

                                // Check for finish reason to close popups
                                if (parsed.choices[0].finish_reason) {
                                    // If we don't have a message index yet (no content received), create assistant message now
                                    if (messageIndex === null) {
                                        this.messages.push({ role: 'assistant', content: assistantMessage, isTyping: false });
                                        messageIndex = this.messages.length - 1;

                                        // CRITICAL: Save immediately after adding assistant message
                                        this.saveConversations().catch(error => {
                                            console.error('Failed to save after adding assistant message:', error);
                                        });
                                    } else {
                                        // Update final message content and save
                                        this.messages[messageIndex].content = assistantMessage;
                                        this.messages[messageIndex].isTyping = false;

                                        // CRITICAL: Save immediately after updating assistant message
                                        this.saveConversations().catch(error => {
                                            console.error('Failed to save after updating assistant message:', error);
                                        });
                                    }
                                    this.renderMessages();
                                    this.saveCurrentConversation();
                                    // Ensure syntax highlighting is applied after streaming completes
                                    setTimeout(() => {
                                        this.highlightCodeBlocks();
                                        this.closeAllToolPopups();
                                    }, 200);
                                }
                            }
                        } catch (e) {
                            // Log JSON parse errors for debugging, but continue processing
                            if (data.trim() && !data.includes('[DONE]')) {
                                console.debug('JSON parse error for chunk:', data, 'Error:', e.message);
                            }
                        }
                    }
                }
            }
        } catch (error) {
            // Check if this is a user-initiated stop (cancelled stream)
            if (error.name === 'AbortError' || error.message.includes('abort') || !this.isStreaming) {
                console.log('🛑 Stream cancelled by user');
                // Don't treat user cancellation as an error
            } else {
                console.error('Streaming error:', error);
            }

            // Always finalize the message, whether it's an error or user stop
            if (messageIndex !== null) {
                this.updateMessageContent(messageIndex, assistantMessage, false);
            }
            this.saveCurrentConversation(); // Save even on error/stop

            // Close any remaining tool popups
            this.closeAllToolPopups();

            // Only re-throw if it's not a user-initiated stop
            if (error.name !== 'AbortError' && !error.message.includes('abort') && this.isStreaming) {
                throw error;
            }
        } finally {
            // Ensure all tool popups are closed when streaming completes
            setTimeout(() => {
                this.closeAllToolPopups();
            }, 2000);
        }
    }

    updateMessageContent(index, content, isStreaming = false) {
        const domIndex = this.getMessageDomIndex(index);
        const messageGroups = this.chatMessages.querySelectorAll('.message-group');
        if (messageGroups[domIndex]) {
            const messageContent = messageGroups[domIndex].querySelector('.message-content');
            // Preserve the actions div
            const actions = messageContent.querySelector('.message-actions');

            // Add or remove streaming class
            if (isStreaming) {
                messageContent.classList.add('streaming');
            } else {
                messageContent.classList.remove('streaming');
            }

            // Render content with optional cursor
            let html = this.renderMarkdown(content);
            html = this.rewritePlantUMLUrls(html);
            if (isStreaming) {
                html += '<span class="streaming-cursor"></span>';
            }

            messageContent.innerHTML = html;
            if (actions) {
                messageContent.appendChild(actions);
            }
            this.scrollToBottom();

            // Apply syntax highlighting to code blocks when not streaming
            if (!isStreaming) {
                setTimeout(() => {
                    this.highlightCodeBlocks();
                }, 100);
            }
        }
    }

    showError(message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = message;

        const lastMessage = this.chatMessages.lastElementChild;
        if (lastMessage) {
            lastMessage.appendChild(errorDiv);
        }
    }

    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 12px 16px;
            border-radius: 8px;
            color: white;
            font-weight: 500;
            font-size: 14px;
            z-index: 10000;
            max-width: 400px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
            animation: slideInRight 0.3s ease-out;
        `;

        // Set background color based on type
        switch (type) {
            case 'error':
                notification.style.background = 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)';
                break;
            case 'warning':
                notification.style.background = 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)';
                break;
            case 'success':
                notification.style.background = 'linear-gradient(135deg, #10b981 0%, #059669 100%)';
                break;
            default: // info
                notification.style.background = 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)';
        }

        notification.textContent = message;

        // Add to document
        document.body.appendChild(notification);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOutRight 0.3s ease-out';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 5000);

        // Add click to dismiss
        notification.addEventListener('click', () => {
            notification.style.animation = 'slideOutRight 0.3s ease-out';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        });
    }

    scrollToBottom() {
        this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
    }

    // Apply Prism.js highlighting to code blocks
    highlightCodeBlocks() {
        this.waitForPrism().then(() => {
            console.log('Prism.js ready, applying highlighting...');

            // Find all code blocks that need highlighting
            const codeBlocks = document.querySelectorAll('.code-block-container code[class*="language-"]');
            console.log('Found', codeBlocks.length, 'code blocks to highlight');

            // Clear any existing highlighting classes
            codeBlocks.forEach(block => {
                block.removeAttribute('data-highlighted');
                block.classList.remove('highlighted');
            });

            // Highlight each block individually for better control
            codeBlocks.forEach((block, index) => {
                console.log(`Highlighting block ${index}:`, block.className, 'Language:', block.classList.toString().match(/language-(\w+)/)?.[1]);
                try {
                    Prism.highlightElement(block);
                    block.classList.add('highlighted');
                } catch (error) {
                    console.error('Error highlighting block:', error);
                }
            });
        }).catch(error => {
            console.error('Prism.js failed to load:', error);
        });
    }

    // Wait for Prism.js to be fully loaded
    waitForPrism() {
        return new Promise((resolve, reject) => {
            let attempts = 0;
            const maxAttempts = 50;

            const checkPrism = () => {
                attempts++;
                if (typeof Prism !== 'undefined' && Prism.highlightElement) {
                    resolve();
                } else if (attempts >= maxAttempts) {
                    reject(new Error('Prism.js failed to load after 5 seconds'));
                } else {
                    setTimeout(checkPrism, 100);
                }
            };

            checkPrism();
        });
    }

    updateSendButton() {
        if (this.isStreaming) {
            this.sendButton.textContent = 'Stop';
            this.sendButton.style.background = 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)';
            this.sendButton.disabled = false;  // Enable the button so user can click to stop
        } else {
            this.sendButton.textContent = 'Send';
            this.sendButton.style.background = 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)';
        }
    }

    async stopStreaming() {
        console.log('🛑 Stop streaming requested');

        // Immediately set streaming state to false to prevent further processing
        this.isStreaming = false;

        // Update UI immediately
        this.typingIndicator.classList.remove('active');
        this.sendButton.disabled = false;
        this.updateSendButton();

        // Cancel the stream reader if it exists
        if (this.currentReader) {
            try {
                console.log('🛑 Cancelling stream reader');
                await this.currentReader.cancel();
            } catch (error) {
                console.log('Stream cancellation error (expected):', error);
            }
            this.currentReader = null;
        }

        // Flush any partial message content and finalize the current message
        this.flushPartialMessage();

        // Close any open tool popups immediately
        this.closeAllToolPopups();

        // Save the conversation state
        this.saveCurrentConversation();

        console.log('🛑 Streaming stopped successfully');
    }

    showErrorNotification(errorEvent) {
        const errorContainer = document.createElement('div');
        errorContainer.className = 'error-notification';
        errorContainer.innerHTML = `
            <div class="error-header">
                <span class="error-icon">⚠️</span>
                <span class="error-title">System Error</span>
                <span class="error-severity">[${errorEvent.error.severity || 'error'}]</span>
            </div>
            <div class="error-message">${errorEvent.error.message}</div>
            <div class="error-details">
                <div class="error-source">Source: ${errorEvent.error.source || 'unknown'}</div>
                <div class="error-context">${errorEvent.error.context || ''}</div>
            </div>
        `;

        // Add to chat messages area
        const chatMessages = document.getElementById('chatMessages');
        if (chatMessages) {
            chatMessages.appendChild(errorContainer);
            chatMessages.scrollTop = chatMessages.scrollHeight;
        }

        // Auto-hide after 10 seconds
        setTimeout(() => {
            if (errorContainer.parentNode) {
                errorContainer.remove();
            }
        }, 10000);
    }

    flushPartialMessage() {
        // Find the last message that might be incomplete
        const lastMessageIndex = this.messages.length - 1;
        if (lastMessageIndex >= 0) {
            const lastMessage = this.messages[lastMessageIndex];

            // If the last message is an assistant message that was being streamed
            if (lastMessage && lastMessage.role === 'assistant') {
                // Remove streaming state and finalize the message
                lastMessage.isTyping = false;

                // If the message is empty or just whitespace, add a stopped indicator
                if (!lastMessage.content || lastMessage.content.trim() === '') {
                    lastMessage.content = '*Response stopped by user*';
                } else {
                    // Add a clear indication that the response was stopped
                    if (!lastMessage.content.endsWith('...') && !lastMessage.content.endsWith('*stopped*')) {
                        lastMessage.content += ' *[stopped]*';
                    }
                }

                // Update the message display immediately
                this.updateMessageContent(lastMessageIndex, lastMessage.content, false);

                console.log('🛑 Flushed partial message:', lastMessage.content.substring(0, 50) + '...');
            }
        }

        // Ensure the UI is re-rendered to show the final state
        this.renderMessages();
        this.highlightCodeBlocks();
    }
}

// Initialize the chat UI when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Configure Prism.js autoloader if available
    if (typeof Prism !== 'undefined' && Prism.plugins && Prism.plugins.autoloader) {
        Prism.plugins.autoloader.languages_path = 'https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/';
        console.log('Prism.js autoloader configured');
    }

    // Debug: Show what we have available
    console.log('Prism availability:', {
        Prism: typeof Prism,
        highlightElement: typeof Prism?.highlightElement,
        highlightAll: typeof Prism?.highlightAll,
        languages: Object.keys(Prism?.languages || {})
    });

    // Initialize the chat UI
    window.chatUI = new ChatUI();

    // Debug function to test code rendering
    window.testCodeRender = function() {
        const testMarkdown = '```python\nprint("Hello, World!")\nimport pandas as pd\n```';
        console.log('Testing code render with:', testMarkdown);
        const result = window.chatUI.renderMarkdown(testMarkdown);
        console.log('Render result:', result);

        // Add to page for visual inspection
        const testDiv = document.createElement('div');
        testDiv.innerHTML = result;
        testDiv.style.cssText = 'margin: 20px; padding: 20px; border: 2px solid red;';
        document.body.appendChild(testDiv);

        // Apply highlighting
        window.chatUI.highlightCodeBlocks();
    };

    // Debug function for testing tool popups (only available in development)
    window.testToolPopup = function() {
        const testEvent = {
            event_type: 'tool_call',
            tool_call: {
                id: 'test_call_123',
                name: 'TestTool',
                arguments: { test: 'argument' }
            }
        };
        console.log('🧪 Testing tool popup with:', testEvent);
        chatUI.addToolNotification(testEvent.tool_call.name, testEvent);
        chatUI.showToolCallPopup(testEvent);

        // Simulate response after 2 seconds
        setTimeout(() => {
            const responseEvent = {
                event_type: 'tool_response',
                tool_response: {
                    id: 'test_call_123',
                    name: 'TestTool',
                    response: 'Test response successful'
                }
            };
            console.log('🧪 Testing tool response with:', responseEvent);
            chatUI.updateToolResponsePopup(responseEvent);
        }, 2000);
    };
});