package prompt

import (
	"fmt"
	"regexp"
	"strings"
)

// PromptService provides prompt enhancement and optimization services
type PromptService struct {
	styleTemplates   map[string]StyleTemplate
	qualityModifiers []string
}

// StyleTemplate defines a template for different artistic styles
type StyleTemplate struct {
	Name        string
	Description string
	Prefix      string
	Suffix      string
	Keywords    []string
}

// PromptAnalysis contains analysis results for a prompt
type PromptAnalysis struct {
	TokenCount     int
	HasSubject     bool
	HasContext     bool
	HasStyle       bool
	HasQualityMods bool
	Suggestions    []string
	Score          int // 0-100
}

// EnhancementOptions configures prompt enhancement
type EnhancementOptions struct {
	AddQualityModifiers bool
	AddStyleKeywords    bool
	StructurePrompt     bool
	TargetAspectRatio   string
	TargetStyle         string
}

// NewPromptService creates a new prompt service with predefined templates
func NewPromptService() *PromptService {
	service := &PromptService{
		styleTemplates: make(map[string]StyleTemplate),
		qualityModifiers: []string{
			"high quality", "detailed", "sharp focus", "professional",
			"4K resolution", "ultra detailed", "masterpiece",
		},
	}
	service.initializeStyleTemplates()
	return service
}

// initializeStyleTemplates sets up predefined style templates
func (ps *PromptService) initializeStyleTemplates() {
	ps.styleTemplates["photographic"] = StyleTemplate{
		Name:        "Photographic",
		Description: "Realistic photographic style",
		Prefix:      "Professional photograph of",
		Suffix:      "shot with DSLR camera, natural lighting",
		Keywords:    []string{"realistic", "photographic", "natural lighting", "DSLR"},
	}

	ps.styleTemplates["artistic"] = StyleTemplate{
		Name:        "Artistic",
		Description: "Digital art style",
		Prefix:      "Digital artwork of",
		Suffix:      "trending on artstation, concept art",
		Keywords:    []string{"digital art", "concept art", "artistic", "stylized"},
	}

	ps.styleTemplates["cinematic"] = StyleTemplate{
		Name:        "Cinematic",
		Description: "Movie-like cinematic style",
		Prefix:      "Cinematic shot of",
		Suffix:      "dramatic lighting, film grain, movie scene",
		Keywords:    []string{"cinematic", "dramatic", "film", "movie"},
	}

	ps.styleTemplates["portrait"] = StyleTemplate{
		Name:        "Portrait",
		Description: "Portrait photography style",
		Prefix:      "Professional portrait of",
		Suffix:      "studio lighting, shallow depth of field",
		Keywords:    []string{"portrait", "headshot", "studio", "professional"},
	}

	ps.styleTemplates["landscape"] = StyleTemplate{
		Name:        "Landscape",
		Description: "Landscape photography style",
		Prefix:      "Scenic landscape of",
		Suffix:      "golden hour lighting, wide angle",
		Keywords:    []string{"landscape", "scenic", "nature", "panoramic"},
	}

	ps.styleTemplates["abstract"] = StyleTemplate{
		Name:        "Abstract",
		Description: "Abstract art style",
		Prefix:      "Abstract representation of",
		Suffix:      "geometric forms, modern art",
		Keywords:    []string{"abstract", "geometric", "modern", "artistic"},
	}
}

// AnalyzePrompt analyzes a prompt and provides feedback
func (ps *PromptService) AnalyzePrompt(prompt string) PromptAnalysis {
	analysis := PromptAnalysis{
		TokenCount:  ps.estimateTokenCount(prompt),
		Suggestions: []string{},
	}

	// Check for basic components
	analysis.HasSubject = ps.hasSubject(prompt)
	analysis.HasContext = ps.hasContext(prompt)
	analysis.HasStyle = ps.hasStyle(prompt)
	analysis.HasQualityMods = ps.hasQualityModifiers(prompt)

	// Generate suggestions
	if !analysis.HasSubject {
		analysis.Suggestions = append(analysis.Suggestions, "Add a clear subject (person, object, or scene)")
	}
	if !analysis.HasContext {
		analysis.Suggestions = append(analysis.Suggestions, "Include contextual details like environment or background")
	}
	if !analysis.HasStyle {
		analysis.Suggestions = append(analysis.Suggestions, "Specify artistic or photographic style")
	}
	if !analysis.HasQualityMods {
		analysis.Suggestions = append(analysis.Suggestions, "Add quality modifiers like 'high quality' or '4K'")
	}
	if analysis.TokenCount > 400 {
		analysis.Suggestions = append(analysis.Suggestions, "Consider shortening prompt (approaching 480 token limit)")
	}

	// Calculate score
	analysis.Score = ps.calculateScore(analysis)

	return analysis
}

// EnhancePrompt improves a basic prompt using enhancement options
func (ps *PromptService) EnhancePrompt(prompt string, options EnhancementOptions) string {
	enhanced := strings.TrimSpace(prompt)

	// Apply style template if specified
	if options.TargetStyle != "" {
		if template, exists := ps.styleTemplates[options.TargetStyle]; exists {
			enhanced = ps.applyStyleTemplate(enhanced, template)
		}
	}

	// Add quality modifiers
	if options.AddQualityModifiers && !ps.hasQualityModifiers(enhanced) {
		enhanced = ps.addQualityModifiers(enhanced)
	}

	// Structure the prompt if requested
	if options.StructurePrompt {
		enhanced = ps.structurePrompt(enhanced)
	}

	// Add aspect ratio specific keywords
	if options.TargetAspectRatio != "" {
		enhanced = ps.addAspectRatioKeywords(enhanced, options.TargetAspectRatio)
	}

	return enhanced
}

// OptimizeForModel optimizes prompt for specific Imagen model
func (ps *PromptService) OptimizeForModel(prompt string, model string) string {
	optimized := prompt

	switch model {
	case "imagen-4.0-ultra-generate-001":
		// Ultra model benefits from detailed descriptions
		optimized = ps.addDetailKeywords(optimized)
		if !strings.Contains(strings.ToLower(optimized), "ultra") {
			optimized = "Ultra detailed " + optimized
		}

	case "imagen-4.0-fast-generate-001":
		// Fast model works better with concise prompts
		optimized = ps.simplifyPrompt(optimized)

	case "imagen-4.0-generate-001":
		// Standard model - balanced approach
		if !ps.hasQualityModifiers(optimized) {
			optimized = ps.addQualityModifiers(optimized)
		}
	}

	return optimized
}

// ValidatePrompt checks if prompt follows best practices
func (ps *PromptService) ValidatePrompt(prompt string) []string {
	var issues []string

	if len(strings.TrimSpace(prompt)) == 0 {
		issues = append(issues, "Prompt cannot be empty")
		return issues
	}

	if ps.estimateTokenCount(prompt) > 480 {
		issues = append(issues, "Prompt exceeds recommended 480 token limit")
	}

	if len(prompt) < 10 {
		issues = append(issues, "Prompt is too short, add more descriptive details")
	}

	// Check for problematic patterns
	if ps.hasVagueTerms(prompt) {
		issues = append(issues, "Avoid vague terms like 'nice', 'good', 'beautiful' - be more specific")
	}

	if ps.hasConflictingStyles(prompt) {
		issues = append(issues, "Prompt contains conflicting style instructions")
	}

	return issues
}

// GetStyleTemplates returns available style templates
func (ps *PromptService) GetStyleTemplates() map[string]StyleTemplate {
	return ps.styleTemplates
}

// Helper methods

func (ps *PromptService) estimateTokenCount(text string) int {
	// Rough estimation: 1 token ≈ 4 characters for English
	return len(text) / 4
}

func (ps *PromptService) hasSubject(prompt string) bool {
	// Simple check for nouns that could be subjects
	subjectKeywords := []string{
		"person", "man", "woman", "child", "people",
		"cat", "dog", "animal", "bird", "tree", "flower",
		"building", "house", "car", "landscape", "portrait",
		"robot", "alien", "dragon", "character",
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, keyword := range subjectKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			return true
		}
	}

	// Check for "of" pattern which often indicates a subject
	return strings.Contains(lowerPrompt, " of ")
}

func (ps *PromptService) hasContext(prompt string) bool {
	contextKeywords := []string{
		"in", "at", "on", "background", "environment", "setting",
		"forest", "city", "beach", "mountain", "room", "studio",
		"outdoor", "indoor", "night", "day", "sunset", "sunrise",
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, keyword := range contextKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			return true
		}
	}
	return false
}

func (ps *PromptService) hasStyle(prompt string) bool {
	styleKeywords := []string{
		"photographic", "artistic", "digital art", "painting",
		"sketch", "watercolor", "oil painting", "realistic",
		"cartoon", "anime", "cinematic", "portrait", "landscape",
		"abstract", "minimalist", "vintage", "modern",
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, keyword := range styleKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			return true
		}
	}
	return false
}

func (ps *PromptService) hasQualityModifiers(prompt string) bool {
	lowerPrompt := strings.ToLower(prompt)
	for _, modifier := range ps.qualityModifiers {
		if strings.Contains(lowerPrompt, strings.ToLower(modifier)) {
			return true
		}
	}
	return false
}

func (ps *PromptService) hasVagueTerms(prompt string) bool {
	vagueTerms := []string{"nice", "good", "beautiful", "pretty", "cool", "awesome"}
	lowerPrompt := strings.ToLower(prompt)

	for _, term := range vagueTerms {
		if strings.Contains(lowerPrompt, term) {
			return true
		}
	}
	return false
}

func (ps *PromptService) hasConflictingStyles(prompt string) bool {
	// Check for conflicting style combinations
	lowerPrompt := strings.ToLower(prompt)

	conflicts := [][]string{
		{"realistic", "cartoon"},
		{"photographic", "abstract"},
		{"vintage", "futuristic"},
	}

	for _, conflict := range conflicts {
		hasFirst := strings.Contains(lowerPrompt, conflict[0])
		hasSecond := strings.Contains(lowerPrompt, conflict[1])
		if hasFirst && hasSecond {
			return true
		}
	}
	return false
}

func (ps *PromptService) calculateScore(analysis PromptAnalysis) int {
	score := 0

	if analysis.HasSubject {
		score += 25
	}
	if analysis.HasContext {
		score += 25
	}
	if analysis.HasStyle {
		score += 25
	}
	if analysis.HasQualityMods {
		score += 15
	}

	// Penalize if too long
	if analysis.TokenCount > 400 {
		score -= 10
	}

	// Bonus for optimal length
	if analysis.TokenCount >= 50 && analysis.TokenCount <= 200 {
		score += 10
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func (ps *PromptService) applyStyleTemplate(prompt string, template StyleTemplate) string {
	// If prompt already starts with the template prefix, don't duplicate
	if strings.HasPrefix(strings.ToLower(prompt), strings.ToLower(template.Prefix)) {
		return prompt
	}

	// Apply template structure
	enhanced := fmt.Sprintf("%s %s", template.Prefix, prompt)

	// Add suffix if it doesn't already contain those elements
	if !ps.containsTemplateElements(prompt, template) {
		enhanced = fmt.Sprintf("%s, %s", enhanced, template.Suffix)
	}

	return enhanced
}

func (ps *PromptService) containsTemplateElements(prompt string, template StyleTemplate) bool {
	lowerPrompt := strings.ToLower(prompt)
	for _, keyword := range template.Keywords {
		if strings.Contains(lowerPrompt, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func (ps *PromptService) addQualityModifiers(prompt string) string {
	// Add a quality modifier at the beginning
	return fmt.Sprintf("High quality, detailed %s", prompt)
}

func (ps *PromptService) structurePrompt(prompt string) string {
	// Simple restructuring - ensure it follows Subject + Context + Style pattern
	words := strings.Fields(prompt)
	if len(words) < 3 {
		return prompt // Too short to restructure meaningfully
	}

	// This is a simplified structure - in practice, you'd use NLP to better identify components
	return prompt // For now, return as-is - more sophisticated structuring would require NLP
}

func (ps *PromptService) addAspectRatioKeywords(prompt string, aspectRatio string) string {
	keywords := map[string]string{
		"16:9": "cinematic, wide shot",
		"9:16": "vertical, portrait orientation",
		"1:1":  "square composition",
		"4:3":  "classic photography ratio",
		"3:4":  "portrait format",
	}

	if keyword, exists := keywords[aspectRatio]; exists {
		if !strings.Contains(strings.ToLower(prompt), strings.ToLower(keyword)) {
			return fmt.Sprintf("%s, %s", prompt, keyword)
		}
	}

	return prompt
}

func (ps *PromptService) addDetailKeywords(prompt string) string {
	detailKeywords := []string{"intricate details", "fine details", "highly detailed"}
	lowerPrompt := strings.ToLower(prompt)

	for _, keyword := range detailKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			return prompt // Already has detail keywords
		}
	}

	return fmt.Sprintf("%s, intricate details", prompt)
}

func (ps *PromptService) simplifyPrompt(prompt string) string {
	// Remove redundant adjectives and simplify for fast model
	simplified := prompt

	// Remove common redundant phrases
	redundant := []string{
		"highly detailed,", "ultra detailed,", "intricate details,",
		"masterpiece,", "trending on artstation,",
	}

	for _, phrase := range redundant {
		simplified = strings.ReplaceAll(simplified, phrase, "")
	}

	// Clean up extra spaces and commas
	re := regexp.MustCompile(`\s*,\s*,\s*`)
	simplified = re.ReplaceAllString(simplified, ", ")
	simplified = strings.TrimSpace(simplified)

	return simplified
}
