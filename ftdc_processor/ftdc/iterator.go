package ftdc

func (it *FTDCDataIterator) NormalisedDocument(includedPatterns []string) map[string]interface{} {
	return normalizeDocument(it.doc, includedPatterns)
}

func (it *FTDCDataIterator) Next() bool {
	for it.it.Next() {
		if it.it.Metadata() != nil {
			it.metadata = it.it.Metadata()
		}
		it.doc = it.it.Document()

		return true
	}

	return false
}
