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

        // Audio recording state
        this.currentAudioSource = 'microphone';
        this.isRecording = false;
        this.isCreatingLap = false;
        this.mediaRecorder = null;
        this.audioStream = null;
        this.recordingStartTime = null;
        this.recordingTimerInterval = null;
        this.audioChunks = [];

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

        // Conversation management (after DOM elements are initialized)
        this.conversations = this.loadConversations();
        this.currentConversationId = null;

        this.init();

        // Initialize conversation after init
        this.initializeConversation();
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
        // First, try to reconstruct markdown from the actual selected DOM elements
        const reconstructedMarkdown = this.reconstructMarkdownFromRange(range);
        if (reconstructedMarkdown) {
            return reconstructedMarkdown;
        }

        // Fallback to the original text-based approach
        // Clean up the selected text for comparison
        const cleanSelection = selectedText.replace(/\s+/g, ' ').trim();

        // If the selection is very small or empty, return as is
        if (cleanSelection.length < 3) {
            return selectedText;
        }

        // Remove action buttons from selection if present
        const cleanSelectionNoButtons = cleanSelection
            .replace(/\s*Edit\s*Replay from here\s*/g, '')
            .replace(/\s*Edit\s*/g, '')
            .replace(/\s*Replay from here\s*/g, '')
            .trim();

        // Try exact substring match first (for simple cases)
        if (markdownContent.includes(cleanSelectionNoButtons)) {
            const startIndex = markdownContent.indexOf(cleanSelectionNoButtons);
            const endIndex = startIndex + cleanSelectionNoButtons.length;
            return markdownContent.substring(startIndex, endIndex);
        }

        // Try direct match with cleaned selection
        if (markdownContent.includes(cleanSelection)) {
            const startIndex = markdownContent.indexOf(cleanSelection);
            const endIndex = startIndex + cleanSelection.length;
            return markdownContent.substring(startIndex, endIndex);
        }

        // Split into words and look for word sequences
        const selectionWords = cleanSelectionNoButtons.split(/\s+/).filter(w => w.length > 0);

        if (selectionWords.length >= 2) {
            // Look for the first few and last few words to find boundaries
            const numWordsToMatch = Math.min(3, Math.floor(selectionWords.length / 2));
            const firstWords = selectionWords.slice(0, numWordsToMatch).join(' ');
            const lastWords = selectionWords.slice(-numWordsToMatch).join(' ');

            const startPos = markdownContent.indexOf(firstWords);
            const endPos = markdownContent.lastIndexOf(lastWords);

            if (startPos >= 0 && endPos >= 0 && endPos >= startPos) {
                return markdownContent.substring(startPos, endPos + lastWords.length).trim();
            }
        }

        // Fallback: try to find any substantial portion of the text
        if (selectionWords.length >= 5) {
            const middleWords = selectionWords.slice(1, -1).join(' ');
            const middlePos = markdownContent.indexOf(middleWords);
            if (middlePos >= 0) {
                // Expand around the middle match
                let start = middlePos;
                let end = middlePos + middleWords.length;

                // Try to expand backwards
                const wordBefore = selectionWords[0];
                const expandedStart = markdownContent.lastIndexOf(wordBefore, start);
                if (expandedStart >= 0 && start - expandedStart < 50) {
                    start = expandedStart;
                }

                // Try to expand forwards
                const wordAfter = selectionWords[selectionWords.length - 1];
                const expandedEnd = markdownContent.indexOf(wordAfter, end);
                if (expandedEnd >= 0 && expandedEnd - end < 50) {
                    end = expandedEnd + wordAfter.length;
                }

                return markdownContent.substring(start, end).trim();
            }
        }

        // If all else fails, return the original selection
        return selectedText;
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
        for (const file of files) {
            if (file.type.startsWith('image/') ||
                file.type === 'application/pdf' ||
                file.type.startsWith('audio/')) {
                try {
                    const dataURL = await this.fileToDataURL(file);
                    this.addFilePreview(dataURL, file.name, file.type);
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

        // Calculate file size
        const fileSize = this.calculateDataURLSize(dataURL);
        const formattedSize = this.formatFileSize(fileSize);

        // Create preview element
        const preview = document.createElement('div');
        preview.className = 'file-preview';
        preview.dataset.fileId = fileData.id;

        if (fileType.startsWith('image/')) {
            // Image preview
            preview.innerHTML = `
                <img src="${dataURL}" alt="${fileName}" title="${fileName}">
                <div class="file-size-badge" style="position: absolute; top: 2px; right: 20px; background: rgba(0,0,0,0.7); color: white; padding: 1px 4px; border-radius: 3px; font-size: 9px; font-weight: 500;">
                    ${formattedSize}
                </div>
                <button class="remove-file" onclick="chatUI.removeFilePreview('${fileData.id}')">×</button>
            `;
        } else if (fileType === 'application/pdf') {
            // PDF preview
            preview.innerHTML = `
                <div class="pdf-icon" title="${fileName}">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
                    </svg>
                    <div style="word-wrap: break-word; font-size: 10px;">${fileName.length > 12 ? fileName.substring(0, 12) + '...' : fileName}</div>
                    <div style="font-size: 8px; color: #666; margin-top: 2px;">${formattedSize}</div>
                </div>
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
            // Request permissions and get audio stream based on selected source
            const stream = await this.getAudioStream();

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

            this.mediaRecorder = new MediaRecorder(stream, options);
            this.audioStream = stream;
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
                // For "both", we'll need to mix streams (simplified approach)
                const micStream = await navigator.mediaDevices.getUserMedia({
                    audio: {
                        echoCancellation: true,
                        noiseSuppression: true,
                        autoGainControl: true
                    }
                });
                // Note: Mixing microphone and system audio requires more complex implementation
                // For now, we'll just use microphone as the primary source
                // Full implementation would require Web Audio API for mixing
                return micStream;

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

            // Stop all tracks in the stream
            if (this.audioStream) {
                this.audioStream.getTracks().forEach(track => track.stop());
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

            // Thresholds for artifact storage: 2MB or 2 minutes (120 seconds)
            const SIZE_THRESHOLD = 2 * 1024 * 1024; // 2MB
            const DURATION_THRESHOLD = 120 * 1000; // 2 minutes in milliseconds

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

            // If this was a lap, start a new recording immediately
            if (this.isCreatingLap) {
                this.isCreatingLap = false;
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
            const response = await fetch(`${this.baseUrl}/artifact`, {
                method: 'POST',
                headers: {
                    'Content-Type': blob.type,
                    'X-Original-Filename': filename
                },
                body: blob
            });

            if (!response.ok) {
                throw new Error(`Artifact upload failed: ${response.status} ${response.statusText}`);
            }

            const result = await response.json();
            console.log('Artifact uploaded successfully:', result.artifactId);
            return result.artifactId;
        } catch (error) {
            console.error('Failed to upload to artifact storage:', error);
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

        // Clean up streams
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
    loadConversations() {
        const saved = localStorage.getItem('chat_conversations');
        return saved ? JSON.parse(saved) : {};
    }

    saveConversations() {
        try {
            localStorage.setItem('chat_conversations', JSON.stringify(this.conversations));
        } catch (error) {
            if (error.name === 'QuotaExceededError') {
                // Try to free up space by removing oldest conversations
                this.cleanupOldConversations();
                // Try saving again
                try {
                    localStorage.setItem('chat_conversations', JSON.stringify(this.conversations));
                } catch (retryError) {
                    console.error('Failed to save even after cleanup:', retryError);
                    // Re-throw the QuotaExceededError so it can be caught by the calling method
                    throw retryError;
                }
            } else {
                throw error;
            }
        }
    }

    cleanupOldConversations() {
        const conversationIds = Object.keys(this.conversations);

        // If we have more than 10 conversations, remove the oldest ones
        if (conversationIds.length > 10) {
            const sortedIds = conversationIds.sort((a, b) =>
                this.conversations[a].lastModified - this.conversations[b].lastModified
            );

            // Remove oldest conversations (keep only 10 most recent)
            const toRemove = sortedIds.slice(0, conversationIds.length - 10);
            toRemove.forEach(id => {
                delete this.conversations[id];
            });

            console.log(`Cleaned up ${toRemove.length} old conversations to free storage space`);
        }

        // Keep all attachment data - no cleanup needed
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

            // Try to save conversations, handle quota exceeded error
            try {
                this.saveConversations();
                this.renderConversationsList();
            } catch (error) {
                if (error.name === 'QuotaExceededError') {
                    console.warn('localStorage quota exceeded, trying to save without attachments');
                    // Try to save without attachments as fallback
                    try {
                        // Temporarily replace messages with sanitized version
                        const originalMessages = this.conversations[this.currentConversationId].messages;
                        this.conversations[this.currentConversationId].messages = this.createMessagesWithoutAttachments();

                        this.saveConversations();
                        this.renderConversationsList();

                        // Show user notification about attachment removal
                        this.showNotification(
                            'Storage quota exceeded. Conversation saved but attachments were removed to save space. Consider clearing old conversations.',
                            'warning'
                        );

                        console.log('Successfully saved conversation without attachments');
                    } catch (fallbackError) {
                        console.error('Failed to save even without attachments:', fallbackError);
                        this.showNotification(
                            'Unable to save conversation: storage quota exceeded even after removing attachments. Please clear old conversations.',
                            'error'
                        );
                    }
                } else {
                    console.error('Error saving conversation:', error);
                    this.showNotification('Error saving conversation: ' + error.message, 'error');
                }
            }
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

                    // Check if this is a large base64 audio (>1MB when base64 decoded)
                    if (audioData.startsWith('data:') && this.calculateDataURLSize(audioData) > 1024 * 1024) {
                        // This is a large audio file - it should have been stored as an artifact
                        // But if we receive it here as base64, we need to handle it gracefully
                        console.warn('Large audio data in message content - consider using artifact storage');
                    }
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

        // Create edit buttons
        const editButtons = document.createElement('div');
        editButtons.className = 'edit-buttons';

        const saveButton = document.createElement('button');
        saveButton.className = 'save-button';
        saveButton.textContent = 'Save';
        saveButton.onclick = () => this.saveEditWithAttachments(index, textarea.value, currentAttachments);

        const cancelButton = document.createElement('button');
        cancelButton.className = 'cancel-button';
        cancelButton.textContent = 'Cancel';
        cancelButton.onclick = () => this.cancelEdit();

        editButtons.appendChild(saveButton);
        editButtons.appendChild(cancelButton);

        editContainer.appendChild(textarea);
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

    manualCleanupStorage() {
        const conversationCount = Object.keys(this.conversations).length;
        if (confirm(`This will remove old conversations to free up storage space. You have ${conversationCount} conversations. Continue?`)) {
            this.cleanupOldConversations();
            this.saveConversations();
            this.renderConversationsList();

            // Flash the cleanup button to indicate success
            const cleanupBtn = document.getElementById('cleanupStorage');
            cleanupBtn.style.background = 'rgba(34, 197, 94, 0.5)';
            cleanupBtn.textContent = 'Done!';
            setTimeout(() => {
                cleanupBtn.style.background = '';
                cleanupBtn.textContent = 'Cleanup';
            }, 1500);
        }
    }

    async getAssistantResponse() {
        this.typingIndicator.classList.add('active');
        this.sendButton.disabled = true;
        this.isStreaming = true;
        this.updateSendButton();  // Update button to show "Stop"

        // Prepare messages with system prompt
        const messagesWithSystem = [
            { role: 'system', content: this.systemPrompt },
            ...this.messages
        ];

        try {
            const response = await fetch(this.apiUrl, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    model: this.buildModelWithTools() || 'gemini-2.0-flash',
                    messages: messagesWithSystem,
                    temperature: 0.7,
                    max_tokens: 2000,
                    stream: true  // Enable streaming
                })
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
                                    } else {
                                        // Update final message content and save
                                        this.messages[messageIndex].content = assistantMessage;
                                        this.messages[messageIndex].isTyping = false;
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